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
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.KubeMeta;
import io.seldon.protos.DeploymentProtos.MLDeployment;

@Component
public class KubeCRDHandlerImpl implements KubeCRDHandler {

	private final static Logger logger = LoggerFactory.getLogger(KubeCRDHandlerImpl.class);
	@Override
	public void updateMLDeployment(DeploymentDef def,CustomResourceDetails crd) {
		
		// Create initial representation of deployment
		MLDeployment mlDeployment = MLDeployment.newBuilder()
				.setApiVersion("machinelearning.seldon.io/v1alpha1")
				.setKind("MLDeployment")
				.setMetadata(KubeMeta.newBuilder()
						.putLabels("app", "seldon")
						.setName(crd.getOref().getName())
						.setNamespace("default")
						.build()
					)
				.setSpec(def)
				.build();
		try
		{
			// Create string representation of JSON to add as annotation to allow declarative "kubectl apply" commands to work otherwise a replcae
			// would remove the last-applied-configuration that kubectl adds.
			String json = JsonFormat.printer().omittingInsignificantWhitespace().preservingProtoFieldNames().print(mlDeployment);
			
			// Create final version of deployment with annotation
			mlDeployment = MLDeployment.newBuilder()
					.setApiVersion("machinelearning.seldon.io/v1alpha1")
					.setKind("MLDeployment")
					.setMetadata(KubeMeta.newBuilder()
							.putLabels("app", "seldon")
							.setName(crd.getOref().getName())
							.setResourceVersion(""+crd.getResourceVersion())
							.setUid(crd.getOref().getUid())
							.putAnnotations("kubectl.kubernetes.io/last-applied-configuration", json+"\n")
							.build()
						)
					.setSpec(def)
					.build();
			
			json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(mlDeployment);
			
			logger.debug(json);
			CustomObjectsApi api = new CustomObjectsApi();
			Object resp = api.replaceNamespacedCustomObject("machinelearning.seldon.io", "v1alpha1", "default", "mldeployments", crd.getOref().getName(),json.getBytes());
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

}
