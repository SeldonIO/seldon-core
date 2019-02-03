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
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.apis.AutoscalingV2beta1Api;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1DeleteOptions;
import io.kubernetes.client.models.V1Service;
import io.kubernetes.client.models.V1ServiceList;
import io.kubernetes.client.models.V1Status;
import io.kubernetes.client.models.V2beta1HorizontalPodAutoscaler;
import io.kubernetes.client.models.V2beta1HorizontalPodAutoscalerList;
import io.kubernetes.client.proto.Meta.DeleteOptions;
import io.kubernetes.client.proto.V1.Service;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.kubernetes.client.proto.V2beta1Autoscaling.HorizontalPodAutoscaler;
import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentControllerImpl implements SeldonDeploymentController {

	private final static Logger logger = LoggerFactory.getLogger(SeldonDeploymentControllerImpl.class);
	private final SeldonDeploymentOperator operator;
	private final K8sClientProvider clientProvider;
	private final KubeCRDHandler crdHandler;
	private final SeldonDeploymentCache mlCache;
	private final SeldonNameCreator seldonNameCreator = new SeldonNameCreator();
	
	private static final String DEPLOYMENT_API_VERSION = "extensions/v1beta1";
	private static final String AUTOSCALER_API_VERSION = "autoscaling/v2beta1";
	
	Cache<String, Boolean> deletedCache = CacheBuilder.newBuilder()
		    .maximumSize(1000)
		    .build();

	
	@Autowired
	public SeldonDeploymentControllerImpl(SeldonDeploymentOperator operator, K8sClientProvider clientProvider,KubeCRDHandler crdHandler,SeldonDeploymentCache mlCache) {
		super();
		this.operator = operator;
		this.clientProvider = clientProvider;
		this.crdHandler = crdHandler;
		this.mlCache = mlCache;
	}
	
	private void createHPAs(ProtoClient client,String namespace,List<HorizontalPodAutoscaler> hpas) throws ApiException, IOException, SeldonDeploymentException
	{
		for(HorizontalPodAutoscaler hpa : hpas)
		{
			 final String listApiPath = "/apis/"+AUTOSCALER_API_VERSION+"/namespaces/{namespace}/deployments/{name}"
	                    .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(hpa.getMetadata().getName()))
	                    .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));
	            logger.debug("Will try to call LIST "+listApiPath);
	            ObjectOrStatus<HorizontalPodAutoscaler> os = client.list(HorizontalPodAutoscaler.newBuilder(),listApiPath);       
	            if (os.status != null) {
	                if (os.status.getCode() == 404) { //Create
	                    logger.debug("About to CREATE "+ProtoBufUtils.toJson(hpa));
	                    final String createApiPath = "/apis/"+AUTOSCALER_API_VERSION+"/namespaces/{namespace}/deployments"
	                            .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));

	                    os = client.create(hpa, createApiPath, AUTOSCALER_API_VERSION, "HorizontalPodAutoscaler");
	                    if (os.status != null) {
	                        logger.error("Error creating HPA:"+ProtoBufUtils.toJson(os.status));
	                        throw new SeldonDeploymentException("Failed to create HPA "+hpa.getMetadata().getName());
	                    }
	                    else {
	                        logger.debug("Created HPA:"+ProtoBufUtils.toJson(os.object));
	                    }
	                }
	                else {
	                    logger.error("Error listing HPAs:"+ProtoBufUtils.toJson(os.status));
	                    throw new SeldonDeploymentException("Failed to list HPA "+hpa.getMetadata().getName());
	                }
	            }
	            else { // Update
	                logger.debug("About to UPDATE "+ProtoBufUtils.toJson(hpa));
	                os = client.update(hpa,listApiPath, AUTOSCALER_API_VERSION, "HorizontalPodAutoscaler");
	                if (os.status != null) {
	                    logger.error("Error updating HPA:"+ProtoBufUtils.toJson(os.status));
	                    throw new SeldonDeploymentException("Failed to update HPA "+hpa.getMetadata().getName());
	                }
	                else {
	                    logger.debug("Updated HPA:"+ProtoBufUtils.toJson(os.object));
	                }
	            }
		}
	}
	
	private void createDeployments(ProtoClient client,String namespace,List<Deployment> deployments) throws ApiException, IOException, SeldonDeploymentException
	{
		for(Deployment d : deployments)
		{
            final String listApiPath = "/apis/"+DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments/{name}"
                    .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(d.getMetadata().getName()))
                    .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));
            logger.debug("Will try to call LIST "+listApiPath);
            ObjectOrStatus<Deployment> os = client.list(Deployment.newBuilder(),listApiPath);       
            if (os.status != null) {
                if (os.status.getCode() == 404) { //Create
                    logger.debug("About to CREATE "+ProtoBufUtils.toJson(d));
                    final String createApiPath = "/apis/"+DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments"
                            .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));

                    os = client.create(d, createApiPath, DEPLOYMENT_API_VERSION, "Deployment");
                    if (os.status != null) {
                        logger.error("Error creating deployment:"+ProtoBufUtils.toJson(os.status));
                        throw new SeldonDeploymentException("Failed to create deployment "+d.getMetadata().getName());
                    }
                    else {
                        logger.debug("Created deployment:"+ProtoBufUtils.toJson(os.object));
                    }
                }
                else {
                    logger.error("Error listing deployment:"+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to list deployment "+d.getMetadata().getName());
                }
            }
            else { // Update
                logger.debug("About to UPDATE "+ProtoBufUtils.toJson(d));
                os = client.update(d,listApiPath, DEPLOYMENT_API_VERSION, "Deployment");
                if (os.status != null) {
                    logger.error("Error updating deployment:"+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to update deployment "+d.getMetadata().getName());
                }
                else {
                    logger.debug("Updated deployment:"+ProtoBufUtils.toJson(os.object));
                }
            }
		}
	}

	private Set<String> getDeploymentNames(List<Deployment> deployments)
	{
		Set<String> names = new HashSet<>();
	    for(Deployment d : deployments)
	        names.add(d.getMetadata().getName());
	    return names;
	}
	
	private int removeDeployments(ProtoClient client,String namespace,SeldonDeployment seldonDeployment,List<Deployment> deployments,boolean svcOrchOnly) throws ApiException, IOException, SeldonDeploymentException
	{
		int deleteCount = 0;
	    Set<String> names = getDeploymentNames(deployments);
	    ExtensionsV1beta1DeploymentList depList = crdHandler.getOwnedDeployments(seldonNameCreator.getSeldonId(seldonDeployment),namespace);
	    for (ExtensionsV1beta1Deployment d : depList.getItems())
	    {
	    	boolean okToDelete = !svcOrchOnly || (d.getMetadata().getLabels().containsKey(Constants.LABEL_SELDON_SVCORCH));
	        if (okToDelete && !names.contains(d.getMetadata().getName()))
	        {
	        	deleteCount++;
	            final String deleteApiPath = "/apis/"+DEPLOYMENT_API_VERSION+"/namespaces/{namespace}/deployments/{name}"
	                    .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(d.getMetadata().getName()))
	                    .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));
	            DeleteOptions options = DeleteOptions.newBuilder().setPropagationPolicy("Foreground").build();
	            ObjectOrStatus<Deployment> os = client.delete(Deployment.newBuilder(),deleteApiPath,options);
	            if (os.status != null) {
                    logger.error("Error deleting deployment:"+ProtoBufUtils.toJson(os.status));
                    //throw new SeldonDeploymentException("Failed to delete deployment "+d.getMetadata().getName());
                }
                else {
                    logger.debug("Deleted deployment:"+ProtoBufUtils.toJson(os.object));
                }
	        }
	        else
	        	logger.info("Skipping deletion of {} svcOrchOnly:{}",d.getMetadata().getName(),svcOrchOnly);
	    }
	    return deleteCount;
	}
	
	private void removeServices(ApiClient client,String namespace,SeldonDeployment seldonDeployment,List<Service> services) throws ApiException, IOException, SeldonDeploymentException
	{
		Set<String> names = getServiceNames(services);
		V1ServiceList svcList = crdHandler.getOwnedServices(seldonNameCreator.getSeldonId(seldonDeployment),namespace);
		for(V1Service s : svcList.getItems())
		{
			if (!names.contains(s.getMetadata().getName()))
			{	
				CoreV1Api api = new CoreV1Api(client);
				io.kubernetes.client.models.V1DeleteOptions options = new V1DeleteOptions();
				V1Status status = api.deleteNamespacedService(s.getMetadata().getName(), namespace, options, null, null, null, null);
				if (!"Success".equals(status.getStatus()))
				{
					logger.error("Failed to delete service "+s.getMetadata().getName());
					throw new SeldonDeploymentException("Failed to delete service "+s.getMetadata().getName());
				}
				else
					logger.debug("Deleted service "+s.getMetadata().getName());
				
			}
		}
	}
	
	private Set<String> getHpaNames(List<HorizontalPodAutoscaler> hpas)
	{
		Set<String> names = new HashSet<>();
	    for(HorizontalPodAutoscaler hpa : hpas)
	        names.add(hpa.getMetadata().getName());
	    return names;
	}
	
	private void removeHPAs(ApiClient client,String namespace,SeldonDeployment seldonDeployment,List<HorizontalPodAutoscaler> hpas) throws ApiException, IOException, SeldonDeploymentException
	{
		Set<String> names = getHpaNames(hpas);
		V2beta1HorizontalPodAutoscalerList hpaList = crdHandler.getOwnedHPAs(seldonNameCreator.getSeldonId(seldonDeployment),namespace);
		for(V2beta1HorizontalPodAutoscaler hpa : hpaList.getItems())
		{
			if (!names.contains(hpa.getMetadata().getName()))
			{	
				AutoscalingV2beta1Api api = new AutoscalingV2beta1Api(client);
				io.kubernetes.client.models.V1DeleteOptions options = new V1DeleteOptions();
				V1Status status = api.deleteNamespacedHorizontalPodAutoscaler(hpa.getMetadata().getName(), namespace, options, null, null, null, null);
				if (!"Success".equals(status.getStatus()))
				{
					logger.error("Failed to delete HPA "+hpa.getMetadata().getName());
					throw new SeldonDeploymentException("Failed to delete HPA "+hpa.getMetadata().getName());
				}
				else
					logger.debug("Deleted HPA "+hpa.getMetadata().getName());
				
			}
		}
	}
	
	/**
	 * Currently Not used as issue with proto client needs further investigation
	 * @param client
	 * @param namespace
	 * @param seldonDeployment
	 * @param services
	 * @throws ApiException
	 * @throws IOException
	 * @throws SeldonDeploymentException
	 */
	private void removeServices(ProtoClient client,String namespace,SeldonDeployment seldonDeployment,List<Service> services) throws ApiException, IOException, SeldonDeploymentException
	{
		Set<String> names = getServiceNames(services);
		V1ServiceList svcList = crdHandler.getOwnedServices(seldonNameCreator.getSeldonId(seldonDeployment),namespace);
		for(V1Service s : svcList.getItems())
		{
			if (!names.contains(s.getMetadata().getName()))
			{	
				final String deleteApiPath = "/apis/v1/namespaces/{namespace}/services/{name}"
	                    .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(s.getMetadata().getName()))
	                    .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));
	            DeleteOptions options = DeleteOptions.newBuilder().setPropagationPolicy("Foreground").build();
	            ObjectOrStatus<Deployment> os = client.delete(Service.newBuilder(),deleteApiPath,options);
	            if (os.status != null) {
                    logger.error("Error deleting service:"+ProtoBufUtils.toJson(os.status));
                    throw new SeldonDeploymentException("Failed to delete service "+s.getMetadata().getName());
                }
                else {
                    logger.debug("Deleted deployment:"+ProtoBufUtils.toJson(os.object));
                }
			}
		}
	}
	
	private Set<String> getServiceNames(List<Service> services)
	{
		Set<String> names = new HashSet<>();
		for(Service s : services)
			names.add(s.getMetadata().getName());
		return names;
	}
	
	private void createServices(ProtoClient client,String namespace,List<Service> services) throws ApiException, IOException, SeldonDeploymentException
	{
		for(Service service : services)
		{
		    final String serviceApiPath = "/api/v1/namespaces/{namespace}/services/{name}"
	                .replaceAll("\\{" + "name" + "\\}", client.getApiClient().escapeString(service.getMetadata().getName()))
	                .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));
	        ObjectOrStatus os = client.list(Service.newBuilder(),serviceApiPath);     
	        if (os.status != null)
	        {
	            if (os.status.getCode() == 404)
	            {
	                String serviceCreateApiPath = "/api/v1/namespaces/{namespace}/services"
	                        .replaceAll("\\{" + "namespace" + "\\}", client.getApiClient().escapeString(namespace));
	                os = client.create(service, serviceCreateApiPath, "v1", "Service");
	                if (os.status != null)
	                {
	                    logger.error("Error creating service "+ProtoBufUtils.toJson(os.status));
	                    throw new SeldonDeploymentException("Failed to create service "+service.getMetadata().getName());
	                }
	                else
	                {
	                    logger.debug("Created service:"+ProtoBufUtils.toJson(os.object));
	                }                   
	            }
	            else
	            {
	                logger.error("Error listing service:"+ProtoBufUtils.toJson(os.status));
	                throw new SeldonDeploymentException("Failed to list service "+service.getMetadata().getName());
	            }
	        }
	        else
	            logger.debug("No creating service as already exists "+service.getMetadata().getName());
		}
	}
	
	private void failDeployment(SeldonDeployment mlDep,Exception e)
	{
        SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder(mlDep);
        mlBuilder.getStatusBuilder().setState(Constants.STATE_FAILED).setDescription(e.getMessage());
        crdHandler.updateSeldonDeploymentStatus(mlBuilder.build());
	}
	
	@Override
	public void removeInitialUnusedResources(SeldonDeployment mlDep) {
		logger.info("Removing initial unused Deployments for Seldon Deployment {}",mlDep.getSpec().getName());
		try
		{
			SeldonDeployment mlDep2 = operator.defaulting(mlDep);
			DeploymentResources resources = operator.createResources(mlDep2);
			ProtoClient client = clientProvider.getProtoClient();
			String namespace = SeldonDeploymentUtils.getNamespace(mlDep2);
			final String deploymentDeleteKey = mlDep.getMetadata().getUid()+":"+mlDep.getMetadata().getResourceVersion();
			logger.info("Deployment delete cache key {}",deploymentDeleteKey);
			if (deletedCache.getIfPresent(deploymentDeleteKey) == null)
			{
				int deleteCount = removeDeployments(client, namespace, mlDep2, resources.deployments,true);
				if (deleteCount == 0)
				{
					logger.info("Failed to delete anything from first stage delete so will delete all unsed deployments for {}",mlDep.getSpec().getName());
					removeDeployments(client, namespace, mlDep2, resources.deployments,false);
				}
				deletedCache.put(deploymentDeleteKey, true);
			}
			else
				logger.info("Skipping initial delete for {}",mlDep.getSpec().getName());
		} catch (SeldonDeploymentException e) {
			logger.error("Failed to cleanup deployment ",e);
		} catch (ApiException e) {
			logger.error("Kubernetes API exception cleaning up code:"+e.getCode()+ "message:"+e.getResponseBody(),e);
		} catch (IOException e) {
			logger.error("IOException during cleanup ",e);
		}
	}

	@Override
	public void removeAllUnusedResources(SeldonDeployment mlDep) {
		logger.info("Removing ALL UNUSED RESOURCES for Seldon Deployment {}",mlDep.getSpec().getName());
		try
		{
			SeldonDeployment mlDep2 = operator.defaulting(mlDep);
			DeploymentResources resources = operator.createResources(mlDep2);
			ProtoClient client = clientProvider.getProtoClient();
			String namespace = SeldonDeploymentUtils.getNamespace(mlDep2);
			removeDeployments(client, namespace, mlDep2, resources.deployments,false);
			ApiClient client2 = clientProvider.getClient();
			removeServices(client2,namespace, mlDep2, resources.services);
			removeHPAs(client2, namespace, mlDep2, resources.hpas);
		} catch (SeldonDeploymentException e) {
			logger.error("Failed to cleanup deployment ",e);
		} catch (ApiException e) {
			logger.error("Kubernetes API exception cleaning up code:"+e.getCode()+ "message:"+e.getResponseBody(),e);
		} catch (IOException e) {
			logger.error("IOException during cleanup ",e);
		}
	}

	
	@Override
	public void createOrReplaceSeldonDeployment(SeldonDeployment mlDep) {

	    if (mlDep.hasStatus() && mlDep.getStatus().hasState() && mlDep.getStatus().getState().equals(Constants.STATE_FAILED))
	    {
	        logger.warn("Ignoring failed deployment "+mlDep.getMetadata().getName());
	        return;
	    }
		try
		{
	        String namespace = SeldonDeploymentUtils.getNamespace(mlDep);
		    SeldonDeployment existing = mlCache.get(mlDep);
		    if (existing == null || !existing.getSpec().equals(mlDep.getSpec()))
		    {
		        logger.debug("Running updates for "+mlDep.getMetadata().getName());
		        mlCache.put(mlDep);
		        SeldonDeployment mlDepStatusUpdated = operator.updateStatus(mlDep);
		        SeldonDeployment mlDep2 = operator.defaulting(mlDep);
		        operator.validate(mlDep2);
		        DeploymentResources resources = operator.createResources(mlDep2);
		        ProtoClient client = clientProvider.getProtoClient();
		        createDeployments(client, namespace, resources.deployments);
		        createServices(client, namespace, resources.services);
		        createHPAs(client, namespace, resources.hpas);
		        if (!mlDep.hasStatus())
		        {
		           logger.debug("Pushing updated SeldonDeployment "+mlDepStatusUpdated.getMetadata().getName()+" back to kubectl");
		           crdHandler.updateSeldonDeploymentStatus(mlDepStatusUpdated);
		        }
		        else
		            logger.debug("Not pushing an update as no change to status for SeldonDeployment "+mlDep2.getMetadata().getName());
		    }
		    else
		    {
		        mlCache.put(mlDep);
		        logger.debug("Only updated cache for "+mlDep.getMetadata().getName());
		    }
			
		} catch (SeldonDeploymentException e) {
			logger.error("Failed to create deployment ",e);
			failDeployment(mlDep,e);
		} catch (ApiException e) {
			logger.error("Kubernetes API exception deploying code:"+e.getCode()+ "message:"+e.getResponseBody(),e);
			failDeployment(mlDep,e);
		} catch (IOException e) {
			logger.error("IOException during createReplace ",e);
			failDeployment(mlDep,e);
		}
	}

	
}
