package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.net.SocketTimeoutException;
import java.text.SimpleDateFormat;
import java.util.Date;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.google.gson.reflect.TypeToken;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.AppsV1beta1Api;
import io.kubernetes.client.models.AppsV1beta1Deployment;
import io.kubernetes.client.models.V1OwnerReference;
import io.kubernetes.client.util.Config;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.k8s.DeploymentUtils.ServiceSelectorDetails;

@Component
public class DeploymentWatcher {

	protected static Logger logger = LoggerFactory.getLogger(DeploymentWatcher.class.getName());
	
	private int resourceVersion = 0;
	private int resourceVersionProcessed = 0;
	
	private final MLDeploymentStatusUpdater statusUpdater;
	
	@Autowired
	public DeploymentWatcher(MLDeploymentStatusUpdater statusUpdater) throws IOException
	{
		this.statusUpdater = statusUpdater;
	}
	
	public int watchDeployments(int resourceVersion,int resourceVersionProcessed) throws ApiException, IOException 
	{
		String rs = null;
		if (resourceVersion > 0)
			rs = ""+resourceVersion;
		logger.info("Watching with rs "+rs);
		
		int maxResourceVersion = resourceVersion;		
		try{
			ApiClient client = Config.defaultClient();
			AppsV1beta1Api api = new AppsV1beta1Api(client);

			//TODO can we use labelSelector to limit to seldon resources
			Watch<AppsV1beta1Deployment> watch = Watch.createWatch(
	                client,
	        		api.listNamespacedDeploymentCall("default", null, null, ServiceSelectorDetails.seldonLabelName+"="+ServiceSelectorDetails.seldonLabelMlDepValue, rs, 10, true,null,null),
	        		new TypeToken<Watch.Response<AppsV1beta1Deployment>>(){}.getType());

			for (Watch.Response<AppsV1beta1Deployment> item : watch) {
				logger.info(String.format("%s\n : %s %d%n", item.type, item.object.getMetadata().getName(),item.object.getStatus().getReadyReplicas()));
	    	
				int resourceVersionNew = Integer.parseInt(item.object.getMetadata().getResourceVersion());
				if (resourceVersionNew <= resourceVersionProcessed)
				{
					logger.warn("Looking at already processed request - skipping");
				}
				else
				{
					if (resourceVersionNew > maxResourceVersion)
						maxResourceVersion = resourceVersionNew;
					switch(item.type)
					{
					case "MODIFIED":
						for (V1OwnerReference ownerRef : item.object.getMetadata().getOwnerReferences())
						{
							if (ownerRef.getKind().equals(KubeCRDHandlerImpl.KIND) && item.object.getStatus() != null && item.object.getStatus().getReadyReplicas() != null)
							{
								String mlDepName = ownerRef.getName();
								String depName = item.object.getMetadata().getName();
								statusUpdater.updateStatus(mlDepName, depName, item.object.getStatus().getReadyReplicas());
							}
						}
						break;
					case "ADDED":
					case "DELETED":
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
		catch(RuntimeException e)
		{
			if (e.getCause() instanceof SocketTimeoutException)
				return maxResourceVersion;
			else
				throw e;
		} 
		return maxResourceVersion;
	}
		
	private static final SimpleDateFormat dateFormat = new SimpleDateFormat("HH:mm:ss");
	
	@Scheduled(fixedDelay = 5000)
    public void watch() throws JsonProcessingException, ApiException, IOException {
        logger.info("The time is now {}", dateFormat.format(new Date()));
        this.resourceVersion = this.watchDeployments(this.resourceVersion,this.resourceVersionProcessed);
        if (this.resourceVersion > this.resourceVersionProcessed)
        {
        	logger.info("Updating processed resource version to "+resourceVersion);
        	this.resourceVersionProcessed = this.resourceVersion;
        }
        else
        {
        	logger.info("Not updating resourceVersion - current:"+this.resourceVersion+" Processed:"+this.resourceVersionProcessed);
        }
    }
		
}
