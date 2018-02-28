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

import org.apache.commons.lang.StringUtils;
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
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentWatcher  {
	protected static Logger logger = LoggerFactory.getLogger(SeldonDeploymentWatcher.class.getName());
	
	private final SeldonDeploymentController seldonDeploymentController;
	private final SeldonDeploymentCache mlCache;
	private final ClusterManagerProperites clusterManagerProperites;
	
	private int resourceVersion = 0;
	private int resourceVersionProcessed = 0;
	
	@Autowired
	public SeldonDeploymentWatcher(ClusterManagerProperites clusterManagerProperites,SeldonDeploymentController seldonDeploymentController,SeldonDeploymentCache mlCache) throws IOException
	{
		this.seldonDeploymentController = seldonDeploymentController;
		this.mlCache = mlCache;
		this.clusterManagerProperites = clusterManagerProperites;
	}
	
	private void processWatch(SeldonDeployment mldep,String action) throws InvalidProtocolBufferException
	{
		switch(action)
		{
		case "ADDED":
		case "MODIFIED":
			seldonDeploymentController.createOrReplaceSeldonDeployment(mldep);
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
		ApiClient client = Config.defaultClient();
		CustomObjectsApi api = new CustomObjectsApi(client);
		String namespace = StringUtils.isEmpty(this.clusterManagerProperites.getNamespace()) ? "default" : this.clusterManagerProperites.getNamespace();
		logger.debug("Watching with rs "+rs+" in namespace "+namespace);
		Watch<Object> watch = Watch.createWatch(
				client,
                api.listNamespacedCustomObjectCall("machinelearning.seldon.io", "v1alpha1", namespace, "seldondeployments", null, null, rs, true, null, null),
                new TypeToken<Watch.Response<Object>>(){}.getType());
		
		int maxResourceVersion = resourceVersion;
		try{
        for (Watch.Response<Object> item : watch) {
        	Gson gson = new GsonBuilder().create();
    		String jsonInString = gson.toJson(item.object);
	    	logger.debug(String.format("%s\n : %s%n", item.type, jsonInString));
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
    	    		
    	    		try
    	    		{
    	    		    this.processWatch(SeldonDeploymentUtils.jsonToSeldonDeployment(jsonInString), item.type);
    	    		}
    	    		catch (InvalidProtocolBufferException e)
    	    		{
    	    		    //TODO : update status of seldondeployment to show error
    	    		    logger.warn("Failed to parse SeldonDelployment " + jsonInString, e);
    	    		}
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
        this.resourceVersion = this.watchSeldonMLDeployments(this.resourceVersion,this.resourceVersionProcessed);
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
