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
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.AppsV1beta1Deployment;
import io.kubernetes.client.models.V1OwnerReference;
import io.kubernetes.client.util.Config;
import io.kubernetes.client.util.Watch;

@Component
public class DeploymentWatcher {

	protected static Logger logger = LoggerFactory.getLogger(DeploymentWatcher.class.getName());
	
	private int resourceVersion = 0;
	private int resourceVersionProcessed = 0;
	
	private final SeldonDeploymentStatusUpdate statusUpdater;
	
	@Autowired
	public DeploymentWatcher(SeldonDeploymentStatusUpdate statusUpdater)
	{
		this.statusUpdater = statusUpdater;
	}
	
	public int watchDeployments(int resourceVersion,int resourceVersionProcessed) throws ApiException, IOException 
	{
		String rs = null;
		if (resourceVersion > 0)
			rs = ""+resourceVersion;
		logger.debug("Watching with rs "+rs);
		
		int maxResourceVersion = resourceVersion;		

		ApiClient client = Config.defaultClient();
		ExtensionsV1beta1Api api = new ExtensionsV1beta1Api(client);

		Watch<AppsV1beta1Deployment> watch = Watch.createWatch(
		        client,
		        api.listNamespacedDeploymentCall("default", null, null, null,false,SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_KEY+"="+SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_VAL, null,rs, 10, true,null,null),
		        new TypeToken<Watch.Response<AppsV1beta1Deployment>>(){}.getType());

		try
		{
		    for (Watch.Response<AppsV1beta1Deployment> item : watch) {
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
                                    statusUpdater.updateStatus(mlDepName, depName, item.object.getStatus().getReplicas(),item.object.getStatus().getReadyReplicas());
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
                                    statusUpdater.removeStatus(mlDepName,depName);
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
