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

import io.fabric8.kubernetes.api.model.OwnerReference;
import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.Configuration;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.util.Config;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

@Component
public class DeploymentWatcher  {
	protected static Logger logger = LoggerFactory.getLogger(DeploymentWatcher.class.getName());
	
	@Autowired
	private ClusterManager clusterManager;

	private ApiClient client;
	private int resourceVersion = 0;
	
	public DeploymentWatcher() throws IOException
	{
		this.client = Config.defaultClient();
		Configuration.setDefaultApiClient(client);
	}
	
	private void processWatch(DeploymentDef deploymentDef,OwnerReference oref,int resourceVersion,String action) throws InvalidProtocolBufferException
	{
		if (action.equals("ADDED"))
		{
			CMResultDef cmResultDef =  clusterManager.createSeldonDeployment(deploymentDef,resourceVersion,oref);
			String json = ProtoBufUtils.toJson(cmResultDef);
			logger.info(json);
		}
		else if (action.equals("MODIFIED"))
			clusterManager.updateSeldonDeployment(deploymentDef,resourceVersion,oref);
		else if (action.equals("DELETED"))
		{
			//TODO check it works with garbage collection in kubernetes 1.8+
			//clusterManager.deleteSeldonDeployment(deploymentDef);
		}
		else
			logger.error("Unknown action "+action);
	}
	
	public int watchSeldonMLDeployments(int resourceVersion) throws ApiException, JsonProcessingException, IOException
	{
		String rs = null;
		if (resourceVersion > 0)
			rs = ""+resourceVersion;
		logger.info("Watching with rs "+rs);
		CustomObjectsApi api = new CustomObjectsApi();
		Watch<Object> watch = Watch.createWatch(
                client,
                api.listNamespacedCustomObjectCall("machinelearning.seldon.io", "v1alpha1", "default", "mldeployments", null, null, rs, true, null, null),
                new TypeToken<Watch.Response<Object>>(){}.getType());
		
		int maxResourceVersion = resourceVersion;
		try{
        for (Watch.Response<Object> item : watch) {
        	Gson gson = new GsonBuilder().setPrettyPrinting().create();
    		String jsonInString = gson.toJson(item.object);
	    	logger.info(String.format("%s\n : %s%n", item.type, jsonInString));
    		ObjectMapper mapper = new ObjectMapper();
    	    JsonFactory factory = mapper.getFactory();
    	    JsonParser parser = factory.createParser(jsonInString);
    	    JsonNode actualObj = mapper.readTree(parser);
    	    if (actualObj.has("kind") && actualObj.get("kind").asText().equals("Status"))
    	    {
    	    	//Issue with resource version - add 1
    	    	logger.warn("Possible old resource version found");
    	    	return maxResourceVersion + 1;
    	    }
    	    else
    	    {
    	    	int resourceVersionNew = actualObj.get("metadata").get("resourceVersion").asInt();
    	    	if (resourceVersionNew > maxResourceVersion)
    	    		maxResourceVersion = resourceVersionNew;
    	    	OwnerReference oref = new OwnerReference(
    	    				actualObj.get("apiVersion").asText(), 
    	    				true, 
    	    				actualObj.get("kind").asText(), 
    	    				actualObj.get("metadata").get("name").asText(), 
    	    				actualObj.get("metadata").get("uid").asText());
    	    	JsonNode deploymentSpec = actualObj.get("spec");
    	    	String jsonDeploymentSpec = mapper.writeValueAsString(deploymentSpec);
    	    	logger.info("Resource version "+resourceVersionNew);
    	    	logger.info(String.format("%s",jsonDeploymentSpec));
    	    	
    	    	 DeploymentDef.Builder deploymentDefBuilder = DeploymentDef.newBuilder();
    	    	 ProtoBufUtils.updateMessageBuilderFromJson(deploymentDefBuilder, jsonDeploymentSpec);
    	    	
    	    	this.processWatch(deploymentDefBuilder.build(), oref,resourceVersionNew, item.type);
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
        this.resourceVersion = this.watchSeldonMLDeployments(this.resourceVersion);
    }

	

}
