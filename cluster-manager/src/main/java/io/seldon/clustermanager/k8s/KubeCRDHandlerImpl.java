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

import java.util.Optional;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.UninitializedMessageException;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.V1ServiceList;
import io.kubernetes.client.models.V2beta1HorizontalPodAutoscalerList;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.kubernetes.client.util.Config;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.k8s.client.K8sApiProvider;
import io.seldon.clustermanager.k8s.client.K8sClientProvider;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

/**
 * Reference implementation for interacting with Seldon Deployments
 * @author clive
 *
 */
@Component
public class KubeCRDHandlerImpl implements KubeCRDHandler {

	private final static Logger logger = LoggerFactory.getLogger(KubeCRDHandlerImpl.class);
	
	public static final String GROUP = "machinelearning.seldon.io";
	public static final String KIND_PLURAL = "seldondeployments";
	public static final String KIND = "SeldonDeployment";
	
	private final boolean clusterWide;
	
	private boolean replaceStatusResource = true; // Whether to use the status CR endpoint (available from k8s 1.10 (alpha) 1.11 (beta)
	
	private final K8sClientProvider k8sClientProvider;
	private final K8sApiProvider k8sApiProvider;
	
	@Autowired
    public KubeCRDHandlerImpl(K8sApiProvider k8sApiProvider,K8sClientProvider k8sClientProvider,ClusterManagerProperites clusterManagerProperites) {
		this.k8sClientProvider= k8sClientProvider;
		this.k8sApiProvider = k8sApiProvider;
		//this.namespace = StringUtils.isEmpty(clusterManagerProperites.getNamespace()) ? "default" : clusterManagerProperites.getNamespace();
		this.clusterWide = !clusterManagerProperites.isSingleNamespace();
		logger.info("Starting with cluster wide {}",clusterWide);
	}
	
	private String getNamespace(SeldonDeployment d)
	{
	    if (StringUtils.isEmpty(d.getMetadata().getNamespace()))
	        return "default";
	    else
	        return d.getMetadata().getNamespace();
	}
	
	
	@Override
	public void updateRaw(String json,String seldonDeploymentName, String version, String namespace) {
		try
		{
			logger.info(json);
			ApiClient client = Config.defaultClient();
			CustomObjectsApi api = new CustomObjectsApi(client);
			if (replaceStatusResource)
			{
				try 
				{
					api.replaceNamespacedCustomObjectStatus(GROUP, version, namespace, KIND_PLURAL, seldonDeploymentName,json.getBytes());
				} catch (ApiException e) {
					replaceStatusResource = false; // Stop using the /status endpoint (maybe because the k8s version does not have this <1.10)
					logger.warn("Failed to update deployment in kubernetes ",e);
				}
			}
			if (!replaceStatusResource)
				api.replaceNamespacedCustomObject(GROUP, version, namespace, KIND_PLURAL, seldonDeploymentName,json.getBytes());
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
			
			logger.debug("Updating seldondeployment {} with status {}",mlDeployment.getMetadata().getName(),mlDeployment.getStatus());
			ApiClient client = k8sClientProvider.getClient();
			CustomObjectsApi api = k8sApiProvider.getCustomObjectsApi(client);
			String namespace = getNamespace(mldep);

			if (replaceStatusResource)
			{
				final String version = SeldonDeploymentUtils.getVersionFromApiVersion(mldep.getApiVersion());
				try 
				{

					api.replaceNamespacedCustomObjectStatus(GROUP, version, namespace, KIND_PLURAL, mlDeployment.getMetadata().getName(),json.getBytes());
				} catch (ApiException e) {
					replaceStatusResource = false; // Stop using the /status endpoint (maybe because the k8s version does not have this <1.10)
					logger.warn("Failed to update deployment in kubernetes {} {}",version,e.getResponseBody(),e);
				}
			}
			if (!replaceStatusResource)
				api.replaceNamespacedCustomObject(GROUP, SeldonDeploymentUtils.getVersionFromApiVersion(mldep.getApiVersion()), namespace, KIND_PLURAL, mlDeployment.getMetadata().getName(),json.getBytes());
		} catch (InvalidProtocolBufferException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (ApiException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (IOException e) {
			logger.error("Failed to get client ",e);
		}
	
	}
	
	@Override
	public SeldonDeployment getSeldonDeployment(String name,String version, String namespace) {
		try
		{
			ApiClient client = k8sClientProvider.getClient();
			CustomObjectsApi api = k8sApiProvider.getCustomObjectsApi(client);
			Object resp = api.getNamespacedCustomObject(GROUP, version, namespace, KIND_PLURAL, name);
			Gson gson = new GsonBuilder().create();
    		String json = gson.toJson(resp);
    		
    		try {
    			return SeldonDeploymentUtils.jsonToSeldonDeployment(json);
			} catch (InvalidProtocolBufferException e) {
				logger.error("Failed to parse "+json,e);
				return null;
			}
    		catch (UninitializedMessageException e)
    		{
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
    public ExtensionsV1beta1DeploymentList getOwnedDeployments(String seldonDeploymentName,String namespace) {
        try
        {
            ApiClient client = k8sClientProvider.getClient();
            ExtensionsV1beta1Api api = k8sApiProvider.getExtensionsV1beta1Api(client);
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
	public V1ServiceList getOwnedServices(String seldonDeploymentName,String namespace) {
		try
		{
			ApiClient client = k8sClientProvider.getClient();
			io.kubernetes.client.apis.CoreV1Api api = k8sApiProvider.getCoreV1Api(client);
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

	@Override
	public Optional<V2beta1HorizontalPodAutoscalerList> getOwnedHPAs(String seldonDeploymentName, String namespace) {
		try
		{
			ApiClient client = k8sClientProvider.getClient();
			io.kubernetes.client.apis.AutoscalingV2beta1Api api = k8sApiProvider.getAutoScalingApi(client);
			V2beta1HorizontalPodAutoscalerList l = api.listNamespacedHorizontalPodAutoscaler(namespace, null, null, null, false, Constants.LABEL_SELDON_ID+"="+seldonDeploymentName, null, null, null, null);
			return Optional.of(l);
		} catch (IOException e) {
            logger.error("Failed to get HPA list for "+seldonDeploymentName,e);
            return Optional.empty();
        } catch (ApiException e) {
            logger.error("Failed to get HPA list for "+seldonDeploymentName,e);
            return Optional.empty();
        }
	}

	
    
    
	

}
