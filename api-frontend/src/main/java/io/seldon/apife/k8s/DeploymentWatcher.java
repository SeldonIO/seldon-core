package io.seldon.apife.k8s;

import java.io.IOException;
import java.net.SocketTimeoutException;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.HashSet;
import java.util.Set;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
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

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.Configuration;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.util.Config;
import io.kubernetes.client.util.Watch;
import io.seldon.apife.deployments.DeploymentsHandler;
import io.seldon.apife.deployments.DeploymentsListener;

@Component
public class DeploymentWatcher  implements DeploymentsHandler{
	
	protected static Logger logger = LoggerFactory.getLogger(DeploymentWatcher.class.getName());
	
	private ApiClient client;
	private int resourceVersion = 0;
	private final Set<DeploymentsListener> listeners;
	
	public DeploymentWatcher() throws IOException
	{
		this.client = Config.defaultClient();
		this.listeners = new HashSet<>();
		Configuration.setDefaultApiClient(client);
	}
	
	private void processWatch(String json,String action)
	{
		if (action.equals("ADDED"))
			for(DeploymentsListener listener: listeners)
				listener.deploymentAdded(json);
		else if (action.equals("MODIFIED"))
			for(DeploymentsListener listener: listeners)
				listener.deploymentUpdated(json);
		else if (action.equals("DELETED"))
			for(DeploymentsListener listener: listeners)
				listener.deploymentRemoved(json);
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
    	    	JsonNode deploymentSpec = actualObj.get("spec");
    	    	String jsonDeploymentSpec = mapper.writeValueAsString(deploymentSpec);
    	    	logger.info("Resource version "+resourceVersionNew);
    	    	logger.info(String.format("%s",jsonDeploymentSpec));
    	    	this.processWatch(jsonDeploymentSpec, item.type);
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

	

	@Override
	public void addListener(DeploymentsListener listener) {
		logger.info("Adding deployment config listener");
        listeners.add(listener);
	}
}
