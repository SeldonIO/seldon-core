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

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.reflect.TypeToken;
import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.util.Config;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.protos.DeploymentProtos.MLDeployment;
import io.seldon.protos.DeploymentProtos.MLDeploymentStatus;

@Component
public class MLDeploymentWatcher  {
	protected static Logger logger = LoggerFactory.getLogger(MLDeploymentWatcher.class.getName());
	
	private final KubernetesManager kubernetesManager;
	private final MLDeploymentCache mlCache;
	
	private int resourceVersion = 0;
	private int resourceVersionProcessed = 0;
	
	@Autowired
	public MLDeploymentWatcher(KubernetesManager kubernetesManager,MLDeploymentCache mlCache) throws IOException
	{
		this.kubernetesManager = kubernetesManager;
		this.mlCache = mlCache;
	}
	
	private MLDeployment addStatusIfNeeded(MLDeployment mldep)
	{
		if (mldep.hasStatus())
		{
			if (!mldep.getSpec().hasPredictorCanary() && mldep.getStatus().getCanaryReplicasReady() > 0)
				return MLDeployment.newBuilder(mldep).setStatus(MLDeploymentStatus.newBuilder(mldep.getStatus()).setCanaryReplicasReady(0)).build();
			else
				return mldep;
		}
		else
		{
			MLDeployment current = mlCache.get(mldep.getMetadata().getName());
			if (current != null)
				if (!mldep.getSpec().hasPredictorCanary() && current.getStatus().getCanaryReplicasReady() > 0)
					return MLDeployment.newBuilder(mldep).setStatus(MLDeploymentStatus.newBuilder(current.getStatus()).setCanaryReplicasReady(0)).build();
				else
					return MLDeployment.newBuilder(mldep).setStatus(current.getStatus()).build();
			else
				return MLDeployment.newBuilder(mldep).setStatus(MLDeploymentStatus.newBuilder().setCanaryReplicasReady(0).setPredictorReplicasReady(0).build()).build();
		}
	}
	
	private void processWatch(MLDeployment mldep,String action) throws InvalidProtocolBufferException
	{
		switch(action)
		{
		case "ADDED":
		case "MODIFIED":
			MLDeployment mlDepUpdated = addStatusIfNeeded(mldep);
			mlCache.put(mlDepUpdated);
			kubernetesManager.createOrReplaceSeldonDeployment(mlDepUpdated);
			break;
		case "DELETED":
			mlCache.remove(mldep.getMetadata().getName());
			// kubernetes >=1.8 has CRD garbage collection
			logger.info("Resource deleted - ignoring");
			break;
		default:
			logger.error("Unknown action "+action);
		}
	}
	
	
	
	public int watchSeldonMLDeployments(int resourceVersion,int resourceVersionProcessed) throws ApiException, JsonProcessingException, IOException
	{
		String rs = null;
		if (resourceVersion > 0)
			rs = ""+resourceVersion;
		logger.info("Watching with rs "+rs);
		ApiClient client = Config.defaultClient();
		CustomObjectsApi api = new CustomObjectsApi(client);
		Watch<Object> watch = Watch.createWatch(
				client,
                api.listNamespacedCustomObjectCall("machinelearning.seldon.io", "v1alpha1", "default", "mldeployments", null, null, rs, true, null, null),
                new TypeToken<Watch.Response<Object>>(){}.getType());
		
		int maxResourceVersion = resourceVersion;
		try{
        for (Watch.Response<Object> item : watch) {
        	Gson gson = new GsonBuilder().create();
    		String jsonInString = gson.toJson(item.object);
	    	logger.info(String.format("%s\n : %s%n", item.type, jsonInString));
    		ObjectMapper mapper = new ObjectMapper();
    	    JsonFactory factory = mapper.getFactory();
    	    JsonParser parser = factory.createParser(jsonInString);
    	    JsonNode actualObj = mapper.readTree(parser);
    	    if (actualObj.has("kind") && actualObj.get("kind").asText().equals("Status"))
    	    {
    	    	logger.warn("Possible old resource version found - resetting");
    	    	return 0;
    	    }
    	    else
    	    {
    	    	int resourceVersionNew = actualObj.get("metadata").get("resourceVersion").asInt();
    	    	if (resourceVersionNew <= resourceVersionProcessed)
    	    	{
    	    		logger.warn("Looking at already processed request - skipping");
    	    	}
    	    	else
    	    	{
    	    		if (resourceVersionNew > maxResourceVersion)
        	    		maxResourceVersion = resourceVersionNew;

    	    		this.processWatch(MLDeploymentUtils.jsonToMLDeployment(jsonInString), item.type);
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
		return maxResourceVersion;
	}
	
	private static final SimpleDateFormat dateFormat = new SimpleDateFormat("HH:mm:ss");
	
	@Scheduled(fixedDelay = 5000)
    public void watch() throws JsonProcessingException, ApiException, IOException {
        logger.info("The time is now {}", dateFormat.format(new Date()));
        this.resourceVersion = this.watchSeldonMLDeployments(this.resourceVersion,this.resourceVersionProcessed);
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
