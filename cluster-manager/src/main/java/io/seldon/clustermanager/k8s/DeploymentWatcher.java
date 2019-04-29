/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.net.SocketTimeoutException;
import java.text.SimpleDateFormat;
import java.util.Date;

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.google.gson.reflect.TypeToken;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentStatus;
import io.kubernetes.client.models.V1OwnerReference;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.k8s.client.K8sApiProvider;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.clustermanager.k8s.tasks.K8sTaskScheduler;
import io.seldon.clustermanager.k8s.tasks.SeldonDeploymentTaskKey;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

/**
 * Watch deployments created by Seldon Deployments to update the status of the owned Seldon Deployment
 * @author clive
 *
 */
@Component
public class DeploymentWatcher {

	protected static Logger logger = LoggerFactory.getLogger(DeploymentWatcher.class.getName());
	
	private int resourceVersion = 0;
	private int resourceVersionProcessed = 0;
	
	private final K8sTaskScheduler taskScheduler;
	private final SeldonDeploymentStatusUpdate statusUpdater;
	private final K8sClientProvider k8sClientProvider;
	private final K8sApiProvider k8sApiProvider;
	private final String namespace;
	private final boolean clusterWide;
	
	
	/**
	 * Runnable to call the status updater
	 * @author clive
	 *
	 */
	public static class StatusUpdateTask implements Runnable {
		
		private final SeldonDeploymentStatusUpdate statusUpdater;	
		private final String sdepName;
		private final String version;
		private final String namespace;
		private final String depName;
		private final Integer replicas;
		private final Integer replicasReady;
		public StatusUpdateTask(SeldonDeploymentStatusUpdate statusUpdater,String sdepName, String version, String namespace,String depName, Integer replicas,Integer replicasReady) {
				this.statusUpdater = statusUpdater;
				this.sdepName = sdepName;
				this.version = version;
				this.namespace = namespace;
				this.depName = depName;
				this.replicas = replicas;
				this.replicasReady = replicasReady;
		}

		@Override
		public void run() {
			statusUpdater.updateStatus(sdepName, version, depName, replicas, replicasReady, namespace);
		}
	}
	
	/**
	 * Taks to remove Seldon Deployment status when deployment removed
	 * @author clive
	 *
	 */
	public static class StatusRemoveTask implements Runnable {
		
		private final SeldonDeploymentStatusUpdate statusUpdater;	
		private final String sdepName;
		private final String version;
		private final String namespace;
		private final String depName;
		public StatusRemoveTask(SeldonDeploymentStatusUpdate statusUpdater,String sdepName, String version, String namespace,String depName) {
				this.statusUpdater = statusUpdater;
				this.sdepName = sdepName;
				this.version = version;
				this.namespace = namespace;
				this.depName = depName;
		}

		@Override
		public void run() {
			statusUpdater.removeStatus(sdepName, version, depName, namespace);
		}
	}
	
	
	@Autowired
	public DeploymentWatcher(K8sApiProvider k8sApiProvider,K8sClientProvider k8sClientProvider,
			ClusterManagerProperites clusterManagerProperites,SeldonDeploymentStatusUpdate statusUpdater,
			K8sTaskScheduler taskScheduler)
	{
		this.statusUpdater = statusUpdater;
		this.namespace = StringUtils.isEmpty(clusterManagerProperites.getNamespace()) ? "default" : clusterManagerProperites.getNamespace();
		this.clusterWide = !clusterManagerProperites.isSingleNamespace();
		this.k8sClientProvider = k8sClientProvider;
		this.k8sApiProvider = k8sApiProvider;
		this.taskScheduler = taskScheduler;
	}
	
	/**
	 * Watch for owned deployments
	 * @param resourceVersion  - last resource version returned
	 * @param resourceVersionProcessed - last resource version processed
	 * @return the new resource version
	 * @throws ApiException
	 * @throws IOException
	 */
	public int watchDeployments(int resourceVersion,int resourceVersionProcessed) throws ApiException, IOException 
	{
		String rs = null;
		if (resourceVersion > 0)
			rs = ""+resourceVersion;
		logger.debug("Watching with rs "+rs);
		
		int maxResourceVersion = resourceVersion;		

		ApiClient client = k8sClientProvider.getClient();
		ExtensionsV1beta1Api api = k8sApiProvider.getExtensionsV1beta1Api(client);

		Watch<ExtensionsV1beta1Deployment> watch;
		if (this.clusterWide)
		{
			watch = Watch.createWatch(
					client,
					api.listDeploymentForAllNamespacesCall(null, null, false, SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_KEY+"="+SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_VAL, null, null, rs, 10, true, null, null),
					new TypeToken<Watch.Response<ExtensionsV1beta1Deployment>>(){}.getType());
		}
		else
		{
			watch = Watch.createWatch(
					client,
					api.listNamespacedDeploymentCall(namespace, null, null, null,false,SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_KEY+"="+SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_VAL, null,rs, 10, true,null,null),
					new TypeToken<Watch.Response<ExtensionsV1beta1Deployment>>(){}.getType());
		}

		try
		{
		    for (Watch.Response<ExtensionsV1beta1Deployment> item : watch) {
		    	if (item.object == null)
		    	{
		    		logger.warn("Bad watch returned will reset resource version type:{} status:{} ",item.type,item.status.toString());
		    		return 0;
		    	}
                int resourceVersionNew = Integer.parseInt(item.object.getMetadata().getResourceVersion());
                if (resourceVersionNew <= resourceVersionProcessed)
                {
                    logger.warn("Looking at already processed request - skipping");
                }
                else
                {
                    if (resourceVersionNew > maxResourceVersion)
                        maxResourceVersion = resourceVersionNew;
                    if (item.object.getMetadata().getOwnerReferences() == null)
                    {
                        logger.warn("Found possible seldon controlled deployment which has no owner reference. Ignoring.");
                    }
                    else
                    {
                        switch(item.type)
                        {
                        case "ADDED":
                        case "MODIFIED":
                            for (V1OwnerReference ownerRef : item.object.getMetadata().getOwnerReferences())
                            {
                                if (ownerRef.getKind().equals(KubeCRDHandlerImpl.KIND) && item.object.getStatus() != null)
                                {
                                    String mlDepName = ownerRef.getName();
                                    String depName = item.object.getMetadata().getName();
                                    String namespace = StringUtils.isEmpty(item.object.getMetadata().getNamespace()) ? "default" : item.object.getMetadata().getNamespace();
                                    ExtensionsV1beta1DeploymentStatus status = item.object.getStatus();
                                    logger.info("{} {} {} replicas:{} replicasAvailable(ready):{} replicasUnavilable:{} replicasReady(available):{}",item.type,mlDepName,depName,status.getReplicas(),status.getReadyReplicas(),status.getUnavailableReplicas(),status.getAvailableReplicas());
                                    final String version = SeldonDeploymentUtils.getVersionFromApiVersion(ownerRef.getApiVersion());
                                    //statusUpdater.updateStatus(mlDep, depName, item.object.getStatus().getReplicas(),item.object.getStatus().getReadyReplicas());
                                    taskScheduler.submit(new SeldonDeploymentTaskKey(mlDepName, version, namespace), new StatusUpdateTask(statusUpdater, mlDepName, version, namespace, depName, item.object.getStatus().getReplicas(),item.object.getStatus().getReadyReplicas()));
                                }
                            }
                            break;
                        case "DELETED":
                            for (V1OwnerReference ownerRef : item.object.getMetadata().getOwnerReferences())
                            {
                                if (ownerRef.getKind().equals(KubeCRDHandlerImpl.KIND) && item.object.getStatus() != null)
                                {
                                    String mlDepName = ownerRef.getName();
                                    String depName = item.object.getMetadata().getName();
                                    ExtensionsV1beta1DeploymentStatus status = item.object.getStatus();
                                    logger.info("{} {} {} replicas:{} replicasAvailable(ready):{} replicasUnavilable:{} replicasReady(available):{}",item.type,mlDepName,depName,status.getReplicas(),status.getReadyReplicas(),status.getUnavailableReplicas(),status.getAvailableReplicas());
                                    String namespace = StringUtils.isEmpty(item.object.getMetadata().getNamespace()) ? "default" : item.object.getMetadata().getNamespace();
                                    final String version = SeldonDeploymentUtils.getVersionFromApiVersion(ownerRef.getApiVersion());
                                    //statusUpdater.removeStatus(mlDep,depName);
                                    taskScheduler.submit(new SeldonDeploymentTaskKey(mlDepName, version, namespace), new StatusRemoveTask(statusUpdater, mlDepName, version, namespace, depName));
                                }
                            }
                            break;
                        default:
                            logger.error("Unknown type "+item.type);
                        }
                        //for modified get owner reference and determine which predictor it is
                        //get the MLDeployment from API or local cache and update status
                        // put this logic in new class
                    }
                }
            }
        }
        catch(RuntimeException e)
        {
            if (e.getCause() instanceof SocketTimeoutException)
                return maxResourceVersion;
            else
                throw e;
        } 
        finally {
            watch.close();
        }
        return maxResourceVersion;
	}
		
	private static final SimpleDateFormat dateFormat = new SimpleDateFormat("HH:mm:ss");
	
	@Scheduled(fixedDelay = 5000)
    public void watch() throws JsonProcessingException, ApiException, IOException {
        logger.debug("The time is now {}", dateFormat.format(new Date()));
        this.resourceVersion = this.watchDeployments(this.resourceVersion,this.resourceVersionProcessed);
        if (this.resourceVersion > this.resourceVersionProcessed)
        {
        	logger.debug("Updating processed resource version to "+resourceVersion);
        	this.resourceVersionProcessed = this.resourceVersion;
        }
        else
        {
        	logger.debug("Not updating resourceVersion - current:"+this.resourceVersion+" Processed:"+this.resourceVersionProcessed);
        }
    }
		
}
