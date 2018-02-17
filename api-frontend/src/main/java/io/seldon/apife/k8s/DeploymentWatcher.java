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
package io.seldon.apife.k8s;

import java.io.IOException;
import java.net.SocketTimeoutException;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.HashSet;
import java.util.Set;

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParseException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
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
import io.seldon.apife.pb.ProtoBufUtils;
import io.seldon.apife.AppProperties;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class DeploymentWatcher  implements DeploymentsHandler{
	
	protected static Logger logger = LoggerFactory.getLogger(DeploymentWatcher.class.getName());
	
	private ApiClient client;
	private int resourceVersion = 0;
	private int resourceVersionProcessed = 0;
	private final Set<DeploymentsListener> listeners;
	private final String namespace;
	
	@Autowired
	public DeploymentWatcher(AppProperties clusterManagerProperites) throws IOException
	{
		this.client = Config.defaultClient();
		this.listeners = new HashSet<>();
		Configuration.setDefaultApiClient(client);
		this.namespace = StringUtils.isEmpty(clusterManagerProperites.getNamespace()) ? "default" : clusterManagerProperites.getNamespace();
	}
	
	private void processWatch(SeldonDeployment mlDep,String action)
	{
		if (action.equals("ADDED"))
			for(DeploymentsListener listener: listeners)
				listener.deploymentAdded(mlDep);
		else if (action.equals("MODIFIED"))
			for(DeploymentsListener listener: listeners)
				listener.deploymentUpdated(mlDep);
		else if (action.equals("DELETED"))
			for(DeploymentsListener listener: listeners)
				listener.deploymentRemoved(mlDep);
		else
			logger.error("Unknown action "+action);
	}
	
	
	
	
	private String removeCreationTimestampField(String json)
	{
		try
		{
		ObjectMapper mapper = new ObjectMapper();
	    JsonFactory factory = mapper.getFactory();
	    JsonParser parser = factory.createParser(json);
	    JsonNode obj = mapper.readTree(parser);
	    if (obj.has("metadata") && obj.get("metadata").has("creationTimestamp"))
	    {
	    	((ObjectNode) obj.get("metadata")).remove("creationTimestamp");
	    	return mapper.writeValueAsString(obj);
	    }
	    else
	    	return json;
		} catch (JsonParseException e) {
			logger.error("Failed to remove creationTimestamp");
			return json;
		} catch (IOException e) {
			logger.error("Failed to remove creationTimestamp");
			return json;
		}
		
	}
	
	public int watchSeldonSeldonDeployments(int resourceVersion,int resourceVersionProcessed) throws ApiException, JsonProcessingException, IOException
	{
		String rs = null;
		if (resourceVersion > 0)
			rs = ""+resourceVersion;
		logger.info("Watching with rs "+rs);
		CustomObjectsApi api = new CustomObjectsApi();
		Watch<Object> watch = Watch.createWatch(
                client,
                api.listNamespacedCustomObjectCall("machinelearning.seldon.io", "v1alpha1", namespace, "seldondeployments", null, null, rs, true, null, null),
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

    	    		String jsonModified = removeCreationTimestampField(jsonInString);
    	    		SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder();
    	    		ProtoBufUtils.updateMessageBuilderFromJson(mlBuilder, jsonModified);
    	    		
    	    		this.processWatch(mlBuilder.build(), item.type);
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
		 this.resourceVersion = this.watchSeldonSeldonDeployments(this.resourceVersion,this.resourceVersionProcessed);
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

	

	@Override
	public void addListener(DeploymentsListener listener) {
		logger.info("Adding deployment config listener");
        listeners.add(listener);
	}
}
