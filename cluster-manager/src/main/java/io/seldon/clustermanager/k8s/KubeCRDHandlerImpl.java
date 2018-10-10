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

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.databind.JsonNode;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1ServiceList;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.kubernetes.client.util.Config;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class KubeCRDHandlerImpl implements KubeCRDHandler {

	private final static Logger logger = LoggerFactory.getLogger(KubeCRDHandlerImpl.class);
	
	public static final String GROUP = "machinelearning.seldon.io";
	public static final String VERSION = "v1alpha2";
	public static final String KIND_PLURAL = "seldondeployments";
	public static final String KIND = "SeldonDeployment";
	
	private final String namespace;
	
	private boolean replaceStatusResource = true; // Whether to use the status CR endpoint (available from k8s 1.10 (alpha) 1.11 (beta)
	
	@Autowired
    public KubeCRDHandlerImpl(ClusterManagerProperites clusterManagerProperites) {
		this.namespace = StringUtils.isEmpty(clusterManagerProperites.getNamespace()) ? "default" : clusterManagerProperites.getNamespace();
	}
	
	@Override
	public void updateRaw(String json,String seldonDeploymentName) {
		try
		{
			logger.info(json);
			ApiClient client = Config.defaultClient();
			CustomObjectsApi api = new CustomObjectsApi(client);
			api.replaceNamespacedCustomObject(GROUP, VERSION, namespace, KIND_PLURAL, seldonDeploymentName,json.getBytes());
		} catch (InvalidProtocolBufferException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (ApiException e) {
			logger.error("Failed to update deployment in kubernetes : {}",e.getResponseBody());
		} catch (IOException e) {
			logger.error("Failed to get client ",e);
		}
		
		
	}
	
	@Override
	public void updateSeldonDeploymentStatus(SeldonDeployment mldep) {
		
		try
		{
			// Need to remove resourceVersion from the representation used for last-applied-configuration otherwise you will errors subsequently using kubectl
			SeldonDeployment mlDepTmp = SeldonDeployment.newBuilder(mldep).setMetadata(ObjectMeta.newBuilder(mldep.getMetadata()).clearResourceVersion()
					.removeAnnotations("kubectl.kubernetes.io/last-applied-configuration").build()).build();
			// Create string representation of JSON to add as annotation to allow declarative "kubectl apply" commands to work otherwise a replace
			// would remove the last-applied-configuration that kubectl adds.
			String json = SeldonDeploymentUtils.toJson(mlDepTmp,true);
			
			// Create final version of deployment with annotation
			SeldonDeployment mlDeployment = SeldonDeployment.newBuilder(mldep).setMetadata(ObjectMeta.newBuilder(mldep.getMetadata())
						.putAnnotations("kubectl.kubernetes.io/last-applied-configuration", json+"\n")).build();
			json = SeldonDeploymentUtils.toJson(mlDeployment,false);
			
			logger.debug("Updating seldondeployment "+mlDeployment.getMetadata().getName());
			ApiClient client = Config.defaultClient();
			CustomObjectsApi api = new CustomObjectsApi(client);
			if (replaceStatusResource)
			{
				try 
				{
					api.replaceNamespacedCustomObjectStatus(GROUP, VERSION, namespace, KIND_PLURAL, mlDeployment.getMetadata().getName(),json.getBytes());
				} catch (ApiException e) {
					replaceStatusResource = false; // Stop using the /status endpoint (maybe because the k8s version does not have this <1.10)
					logger.error("Failed to update deployment in kubernetes ",e);
				}
			}
			if (!replaceStatusResource)
				api.replaceNamespacedCustomObject(GROUP, VERSION, namespace, KIND_PLURAL, mlDeployment.getMetadata().getName(),json.getBytes());
		} catch (InvalidProtocolBufferException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (ApiException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (IOException e) {
			logger.error("Failed to get client ",e);
		}
	
	}
	
	@Override
	public SeldonDeployment getSeldonDeployment(String name) {
		try
		{
			ApiClient client = Config.defaultClient();
			CustomObjectsApi api = new CustomObjectsApi(client);
			Object resp = api.getNamespacedCustomObject(GROUP, VERSION, namespace, KIND_PLURAL, name);
			Gson gson = new GsonBuilder().create();
    		String json = gson.toJson(resp);
    		
    		try {
    			return SeldonDeploymentUtils.jsonToSeldonDeployment(json);
			} catch (InvalidProtocolBufferException e) {
				logger.error("Failed to parse "+json,e);
				return null;
			}
		} catch (ApiException e) {
			return null;
		} catch (IOException e) {
			logger.error("Failed to get client",e);
			return null;
		} 
	}

    @Override
    public ExtensionsV1beta1DeploymentList getOwnedDeployments(String seldonDeploymentName) {
        try
        {
            ApiClient client = Config.defaultClient();
            ExtensionsV1beta1Api api = new ExtensionsV1beta1Api(client);
            ExtensionsV1beta1DeploymentList l =  api.listNamespacedDeployment(namespace, null, null, null, false, Constants.LABEL_SELDON_ID+"="+seldonDeploymentName, null, null, null, false);
            return l;
        } catch (IOException e) {
            logger.error("Failed to get deployment list for "+seldonDeploymentName,e);
            return null;
        } catch (ApiException e) {
            logger.error("Failed to get deployment list for "+seldonDeploymentName,e);
            return null;
        }
    }

	@Override
	public V1ServiceList getOwnedServices(String seldonDeploymentName) {
		try
		{
			ApiClient client = Config.defaultClient();
			io.kubernetes.client.apis.CoreV1Api api = new CoreV1Api(client);
			V1ServiceList l = api.listNamespacedService(namespace, null, null, null, false, Constants.LABEL_SELDON_ID+"="+seldonDeploymentName, null, null, null, null);
			return l;
		} catch (IOException e) {
            logger.error("Failed to get deployment list for "+seldonDeploymentName,e);
            return null;
        } catch (ApiException e) {
            logger.error("Failed to get deployment list for "+seldonDeploymentName,e);
            return null;
        }
	}

	
    
    
	

}
