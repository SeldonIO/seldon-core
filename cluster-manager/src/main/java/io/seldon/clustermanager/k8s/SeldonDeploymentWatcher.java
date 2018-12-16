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
import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.util.Watch;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.k8s.client.K8sApiProvider;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.clustermanager.pb.JsonFormat;
import io.seldon.clustermanager.pb.JsonFormat.Printer;
import io.seldon.protos.DeploymentProtos.DeploymentStatus;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentWatcher  {
	protected static Logger logger = LoggerFactory.getLogger(SeldonDeploymentWatcher.class.getName());
	
	private final SeldonDeploymentController seldonDeploymentController;
	private final SeldonDeploymentCache mlCache;
	private final ClusterManagerProperites clusterManagerProperites;
	private final KubeCRDHandler crdHandler;
	private final K8sClientProvider k8sClientProvider;
	private final K8sApiProvider k8sApiProvider;
	
	private int resourceVersion = 0;
	private int resourceVersionProcessed = 0;
	
	@Autowired
	public SeldonDeploymentWatcher(K8sApiProvider k8sApiProvider,K8sClientProvider k8sClientProvider,CRDCreator crdCreator,ClusterManagerProperites clusterManagerProperites,SeldonDeploymentController seldonDeploymentController,SeldonDeploymentCache mlCache,KubeCRDHandler crdHandler) throws IOException, ApiException
	{
		this.seldonDeploymentController = seldonDeploymentController;
		this.mlCache = mlCache;
		this.clusterManagerProperites = clusterManagerProperites;
		this.crdHandler = crdHandler;
		this.k8sClientProvider = k8sClientProvider;
		this.k8sApiProvider = k8sApiProvider;
		crdCreator.createCRD();
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
			mlCache.remove(mldep);
			// kubernetes >=1.8 has CRD garbage collection
			logger.info("Resource deleted - ignoring");
			break;
		default:
			logger.error("Unknown action "+action);
		}
	}
	
	private void failDeployment(JsonNode mlDep,Exception e,String namespace)
	{
		try
		{
			//Create status message
			DeploymentStatus.Builder statusBuilder = DeploymentStatus.newBuilder(); 
			statusBuilder.setState(Constants.STATE_FAILED).setDescription(e.getMessage());
			//Get JSON for status message
			Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
			ObjectMapper mapper = new ObjectMapper();
			JsonFactory factory = mapper.getFactory();
			JsonParser parser = factory.createParser(jsonPrinter.print(statusBuilder));
			JsonNode statusObj = mapper.readTree(parser);
			//Update deployment json with status
			((ObjectNode) mlDep).set("status", statusObj);
			String json = mapper.writeValueAsString(mlDep);
			String name = mlDep.get("metadata").get("name").asText();
			//Update seldon deployment
			crdHandler.updateRaw(json, name,namespace);
		} catch (JsonParseException e1) {
			logger.error("Fasile to create status for failed parse",e);
		} catch (InvalidProtocolBufferException e1) {
			logger.error("Fasile to create status for failed parse",e);
		} catch (IOException e1) {
			logger.error("Fasile to create status for failed parse",e);
		}
	}
	
	private String getNamespace(JsonNode actualObj)
	{
		if (!clusterManagerProperites.isSingleNamespace())
		{
			if (actualObj.has("metadata") && actualObj.get("meta").has("namespace"))
				return actualObj.get("metadata").get("namespace").asText();
			else
				return "default";
		}
		else
			return StringUtils.isEmpty(this.clusterManagerProperites.getNamespace()) ? "default" : this.clusterManagerProperites.getNamespace();
	}
	
	public int watchSeldonMLDeployments(int resourceVersion,int resourceVersionProcessed) throws ApiException, JsonProcessingException, IOException
	{
		String rs = null;
		if (resourceVersion > 0)
			rs = ""+resourceVersion;
		ApiClient client = k8sClientProvider.getClient();
		CustomObjectsApi api = k8sApiProvider.getCustomObjectsApi(client);
		Watch<Object> watch;
		if (!clusterManagerProperites.isSingleNamespace())
		{
			watch = Watch.createWatch(
					client,
					api.listClusterCustomObjectCall(KubeCRDHandlerImpl.GROUP, KubeCRDHandlerImpl.VERSION,  KubeCRDHandlerImpl.KIND_PLURAL, null, null, rs, true, null, null),
					new TypeToken<Watch.Response<Object>>(){}.getType());
			
		}
		else
		{
			String namespace = StringUtils.isEmpty(this.clusterManagerProperites.getNamespace()) ? "default" : this.clusterManagerProperites.getNamespace();
			logger.debug("Watching with rs "+rs+" in namespace "+namespace);
			watch = Watch.createWatch(
					client,
					api.listNamespacedCustomObjectCall(KubeCRDHandlerImpl.GROUP, KubeCRDHandlerImpl.VERSION, namespace,  KubeCRDHandlerImpl.KIND_PLURAL, null, null, rs, true, null, null),
					new TypeToken<Watch.Response<Object>>(){}.getType());
		}
		
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
    	    			if ("ADDED".equals(item.type))
    	    			{
    	    				failDeployment(actualObj, e, getNamespace(actualObj));
    	    				logger.warn("Failed to parse SeldonDelployment " + jsonInString, e);
    	    			}
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
