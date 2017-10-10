package io.seldon.apife.k8s;

import java.io.IOException;
import java.net.SocketTimeoutException;
import java.util.Date;

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
import java.text.SimpleDateFormat;

@Component
public class DeploymentWatcher {
	
	protected static Logger logger = LoggerFactory.getLogger(DeploymentWatcher.class.getName());
	
	private ApiClient client;
	private int resourceVersion = 0;
	
	public DeploymentWatcher() throws IOException
	{
		this.client = Config.defaultClient();
		Configuration.setDefaultApiClient(client);
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
    		ObjectMapper mapper = new ObjectMapper();
    	    JsonFactory factory = mapper.getFactory();
    	    JsonParser parser = factory.createParser(jsonInString);
    	    JsonNode actualObj = mapper.readTree(parser);
    	    int resourceVersionNew = actualObj.get("metadata").get("resourceVersion").asInt();
    	    if (resourceVersionNew > maxResourceVersion)
    	    	maxResourceVersion = resourceVersionNew;
    	    logger.info("Resource version "+resourceVersionNew);
        	logger.info(String.format("%s\n : %s%n", item.type, jsonInString));
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
    public void reportCurrentTime() throws JsonProcessingException, ApiException, IOException {
        logger.info("The time is now {}", dateFormat.format(new Date()));
        this.resourceVersion = this.watchSeldonMLDeployments(this.resourceVersion);
    }
}
