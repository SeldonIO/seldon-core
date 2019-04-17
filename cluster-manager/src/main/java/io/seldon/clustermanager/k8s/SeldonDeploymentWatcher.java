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
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

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

/**
 * Watch for Seldon Deployment updates. This is the core task that starts all processing in the Operator.
 * @author clive
 *
 */
@Component
public class SeldonDeploymentWatcher  {
	protected static Logger logger = LoggerFactory.getLogger(SeldonDeploymentWatcher.class.getName());
	
	public static final String[] VERSIONS = {"v1alpha2"};
	
	private final SeldonDeploymentController seldonDeploymentController;
	private final SeldonDeploymentCache mlCache;
	private final ClusterManagerProperites clusterManagerProperites;
	private final KubeCRDHandler crdHandler;
	private final K8sClientProvider k8sClientProvider;
	private final K8sApiProvider k8sApiProvider;
	private final Map<String,ResourceVersion> resourceVersions;	
	
	public static class ResourceVersion {
		public final String version;
		public int resourceVersion = 0;
		public int resourceVersionProcessed = 0;
		public ResourceVersion(String version) {
			this.version = version;
		}
		@Override
		public String toString() {
			return "ResourceVersion [version=" + version + ", resourceVersion=" + resourceVersion
					+ ", resourceVersionProcessed=" + resourceVersionProcessed + "]";
		}
		
	}
	
	@Autowired
	public SeldonDeploymentWatcher(K8sApiProvider k8sApiProvider,K8sClientProvider k8sClientProvider,CRDCreator crdCreator,ClusterManagerProperites clusterManagerProperites,SeldonDeploymentController seldonDeploymentController,SeldonDeploymentCache mlCache,KubeCRDHandler crdHandler) throws IOException, ApiException
	{
		this.seldonDeploymentController = seldonDeploymentController;
		this.mlCache = mlCache;
		this.clusterManagerProperites = clusterManagerProperites;
		this.crdHandler = crdHandler;
		this.k8sClientProvider = k8sClientProvider;
		this.k8sApiProvider = k8sApiProvider;
		this.resourceVersions = new ConcurrentHashMap<>();
		for(String version: VERSIONS)
			resourceVersions.put(version, new ResourceVersion(version));
		crdCreator.createCRD();
	}
	
	private synchronized void processWatch(SeldonDeployment mldep,String action) throws InvalidProtocolBufferException
	{
		switch(action)
		{
		case "ADDED":
			seldonDeploymentController.createOrReplaceSeldonDeployment(mldep,true);
			break;
		case "MODIFIED":
			seldonDeploymentController.createOrReplaceSeldonDeployment(mldep,false);
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
	
	private synchronized void failDeployment(JsonNode mlDep,Exception e,String namespace)
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
			String apiVersion = mlDep.get("apiVersion").asText();
			//Update seldon deployment
			crdHandler.updateRaw(json, name, SeldonDeploymentUtils.getVersionFromApiVersion(apiVersion), namespace);
		} catch (JsonParseException e1) {
			logger.error("Failed to create status for failed parse",e);
		} catch (InvalidProtocolBufferException e1) {
			logger.error("Failed to create status for failed parse",e);
		} catch (IOException e1) {
			logger.error("Failed to create status for failed parse",e);
		}
	}
	
	private String getNamespace(JsonNode actualObj)
	{
		if (!clusterManagerProperites.isSingleNamespace())
		{
			if (actualObj.has("metadata") && actualObj.get("metadata").has("namespace"))
				return actualObj.get("metadata").get("namespace").asText();
			else
				return "default";
		}
		else
			return StringUtils.isEmpty(this.clusterManagerProperites.getNamespace()) ? "default" : this.clusterManagerProperites.getNamespace();
	}
	
	public int watchSeldonMLDeployments(String version, int resourceVersion,int resourceVersionProcessed) throws ApiException, JsonProcessingException, IOException
	{
		logger.debug("Called watch for {}",version);
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
					api.listClusterCustomObjectCall(KubeCRDHandlerImpl.GROUP, version,  KubeCRDHandlerImpl.KIND_PLURAL, null, null, rs, true, null, null),
					new TypeToken<Watch.Response<Object>>(){}.getType());
			
		}
		else
		{
			String namespace = StringUtils.isEmpty(this.clusterManagerProperites.getNamespace()) ? "default" : this.clusterManagerProperites.getNamespace();
			logger.debug("Watching with rs {} in namespace {} for {}",rs,namespace,version);
			watch = Watch.createWatch(
					client,
					api.listNamespacedCustomObjectCall(KubeCRDHandlerImpl.GROUP, version, namespace,  KubeCRDHandlerImpl.KIND_PLURAL, null, null, rs, true, null, null),
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
    	    	logger.warn("Possible old resource version found - resetting {}",version);
    	    	return 0;
    	    }
    	    else
    	    {
    	    	int resourceVersionNew = actualObj.get("metadata").get("resourceVersion").asInt();
    	    	if (resourceVersionNew <= resourceVersionProcessed)
    	    	{
    	    		logger.warn("Looking at already processed request - skipping {} {}",version,resourceVersionNew);
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
    	    				logger.warn("Failed to parse SeldonDeployment {}",jsonInString, e);
    	    			}
    	    		}
    	    	}
    	    }
        }
		}
		catch(RuntimeException e)
		{
			if (e.getCause() instanceof SocketTimeoutException)
			{
				logger.debug("Watch timed out ");
				return maxResourceVersion;
			}
			else
				throw e;
		}
		finally {
		    watch.close();
		}
		return maxResourceVersion;
	}
	
	private static final SimpleDateFormat dateFormat = new SimpleDateFormat("HH:mm:ss");
	
	public void runWatch(String version) throws JsonProcessingException, ApiException, IOException
	{
    	ResourceVersion r = this.resourceVersions.get(version);
    	r.resourceVersion = this.watchSeldonMLDeployments(version, r.resourceVersion,r.resourceVersionProcessed);
    	if (r.resourceVersion > r.resourceVersionProcessed)
    	{
    		logger.debug("Updating processed resource version to {}",r.toString());
    		r.resourceVersionProcessed = r.resourceVersion;
    	}
    	else
    	{
    		logger.debug("Not updating resourceVersion: {}",r.toString());
    	}
    }
	
	@Scheduled(fixedDelay = 5000)
    public void watchv1alpha2() throws JsonProcessingException, ApiException, IOException {
        logger.debug("The time is now {}", dateFormat.format(new Date()));
        final String version = VERSIONS[0];
        runWatch(version);
    }

	/*
	@Scheduled(fixedDelay = 5000)
    public void watchv1alpha3() throws JsonProcessingException, ApiException, IOException {
        logger.debug("The time is now {}", dateFormat.format(new Date()));
        final String version = VERSIONS[1];
        runWatch(version);
    }
    */

	

}
