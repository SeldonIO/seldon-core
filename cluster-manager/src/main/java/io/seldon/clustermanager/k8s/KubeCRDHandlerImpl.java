package io.seldon.clustermanager.k8s;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;

import io.fabric8.kubernetes.api.model.OwnerReference;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CustomObjectsApi;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.KubeMeta;
import io.seldon.protos.DeploymentProtos.MLDeployment;

@Component
public class KubeCRDHandlerImpl implements KubeCRDHandler {

	 private final static Logger logger = LoggerFactory.getLogger(KubeCRDHandlerImpl.class);
	@Override
	public void updateMLDeployment(DeploymentDef def,int resourceVersion,OwnerReference oref) {
		MLDeployment mlDeployment = MLDeployment.newBuilder()
				.setApiVersion("machinelearning.seldon.io/v1alpha1")
				.setKind("MLDeployment")
				.setMetadata(KubeMeta.newBuilder()
						.putLabels("app", "seldon")
						.setName(oref.getName())
						.setResourceVersion(""+resourceVersion)
						.build()
					)
				.setSpec(def)
				.build();
		try
		{
			String json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames()
					.print(mlDeployment);
			logger.info(json);
			CustomObjectsApi api = new CustomObjectsApi();
			Object resp = api.replaceNamespacedCustomObject("machinelearning.seldon.io", "v1alpha1", "default", "mldeployments", oref.getName(),json.getBytes());
			Gson gson = new GsonBuilder().setPrettyPrinting().create();
			String jsonInString = gson.toJson(resp);
			logger.info("%s"+jsonInString);
		} catch (InvalidProtocolBufferException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		} catch (ApiException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		finally{}
	}

}
