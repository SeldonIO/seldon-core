package io.seldon.clustermanager.k8s;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.kubernetes.client.util.Config;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class KubeCRDHandlerImpl implements KubeCRDHandler {

	private final static Logger logger = LoggerFactory.getLogger(KubeCRDHandlerImpl.class);
	
	public static final String GROUP = "machinelearning.seldon.io";
	public static final String VERSION = "v1alpha1";
	//TODO make namespace configurable
	public static final String NAMESPACE = "default";
	public static final String KIND_PLURAL = "mldeployments";
	public static final String KIND = "MLDeployment";
	
	
	@Override
	public void updateSeldonDeployment(SeldonDeployment mldep) {
		
		try
		{
			// Need to remove resourceVersion from the representation used for last-applied-configuration otherwise you will errors subsequently using kubectl
			SeldonDeployment mlDepTmp = SeldonDeployment.newBuilder(mldep).setMetadata(ObjectMeta.newBuilder(mldep.getMetadata()).clearResourceVersion()
					.removeAnnotations("kubectl.kubernetes.io/last-applied-configuration").build()).build();
			// Create string representation of JSON to add as annotation to allow declarative "kubectl apply" commands to work otherwise a replace
			// would remove the last-applied-configuration that kubectl adds.
			String json = JsonFormat.printer().omittingInsignificantWhitespace().preservingProtoFieldNames().print(mlDepTmp);
			
			// Create final version of deployment with annotation
			SeldonDeployment mlDeployment = SeldonDeployment.newBuilder(mldep).setMetadata(ObjectMeta.newBuilder(mldep.getMetadata())
						.putAnnotations("kubectl.kubernetes.io/last-applied-configuration", json+"\n")).build();
			
			json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(mlDeployment);
			
			logger.debug(json);
			ApiClient client = Config.defaultClient();
			CustomObjectsApi api = new CustomObjectsApi(client);
			api.replaceNamespacedCustomObject(GROUP, VERSION, NAMESPACE, KIND_PLURAL, mlDeployment.getMetadata().getName(),json.getBytes());
		} catch (InvalidProtocolBufferException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (ApiException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (IOException e) {
			logger.error("Failed to get client ",e);
		}
		finally{}
	}
	
	@Override
	public SeldonDeployment getSeldonDeployment(String name) {
		try
		{
			ApiClient client = Config.defaultClient();
			CustomObjectsApi api = new CustomObjectsApi(client);
			Object resp = api.getNamespacedCustomObject(GROUP, VERSION, NAMESPACE, KIND_PLURAL, name);
			Gson gson = new GsonBuilder().create();
    		String json = gson.toJson(resp);
    		
    		try {
    			return SeldonDeploymentUtils.jsonToMLDeployment(json);
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
	

}
