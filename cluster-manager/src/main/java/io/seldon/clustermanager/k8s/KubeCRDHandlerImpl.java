package io.seldon.clustermanager.k8s;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;

import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.MLDeployment;

@Component
public class KubeCRDHandlerImpl implements KubeCRDHandler {

	private final static Logger logger = LoggerFactory.getLogger(KubeCRDHandlerImpl.class);
	@Override
	public void updateMLDeployment(MLDeployment mldep) {
		
		try
		{
			// Need to remove resourceVersion from the reprsentation used for last-applied-configuration otherwise you will errors subsequnetly using kubectl
			MLDeployment mlDepTmp = MLDeployment.newBuilder(mldep).setMetadata(ObjectMeta.newBuilder(mldep.getMetadata()).clearResourceVersion().build()).build();
			// Create string representation of JSON to add as annotation to allow declarative "kubectl apply" commands to work otherwise a replace
			// would remove the last-applied-configuration that kubectl adds.
			String json = JsonFormat.printer().omittingInsignificantWhitespace().preservingProtoFieldNames().print(mlDepTmp);
			
			// Create final version of deployment with annotation
			MLDeployment mlDeployment = MLDeployment.newBuilder(mldep).setMetadata(ObjectMeta.newBuilder(mldep.getMetadata())
						.putAnnotations("kubectl.kubernetes.io/last-applied-configuration", json+"\n")).build();
			
			json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(mlDeployment);
			
			logger.debug(json);
			CustomObjectsApi api = new CustomObjectsApi();
			Object resp = api.replaceNamespacedCustomObject("machinelearning.seldon.io", "v1alpha1", "default", "mldeployments", mlDeployment.getMetadata().getName(),json.getBytes());
			Gson gson = new GsonBuilder().setPrettyPrinting().create();
			String jsonInString = gson.toJson(resp);
			logger.debug("Response from kubernetes API:%s"+jsonInString);
		} catch (InvalidProtocolBufferException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		} catch (ApiException e) {
			logger.error("Failed to update deployment in kubernetes ",e);
		}
		finally{}
	}
	@Override
	public DeploymentDef getMlDeployment(MLDeployment mlDep) {
		// TODO Auto-generated method stub
		return null;
	}

}
