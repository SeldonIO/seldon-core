package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;

import org.junit.Ignore;
import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.gson.reflect.TypeToken;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;
import com.google.protobuf.util.JsonFormat.Printer;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.JSON;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.kubernetes.client.proto.V1beta1Extensions;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.kubernetes.client.proto.V1beta1Extensions.DeploymentSpec;
import io.kubernetes.client.util.Config;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.MLDeployment;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class ProtoClientTest extends AppTest {

	private final static Logger logger = LoggerFactory.getLogger(ProtoClientTest.class);
	
	private static String getKubernetesDeploymentId(String deploymentName,String predictorName, boolean isCanary) {
		return "sd-" + deploymentName + "-" + predictorName + "-" + ((isCanary) ? "c" : "p");
	}
	
	@Test 
	public void protoTest() throws IOException, ApiException
	{
		ApiClient apiClient = Config.defaultClient();
		ProtoClient pc = new ProtoClient(apiClient);
		
		MLDeploymentOperator op = new MLDeploymentOperatorImpl(getProps());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		MLDeployment mlDep = MLDeploymentUtils.jsonToMLDeployment(jsonStr);
		mlDep = op.defaulting(mlDep);
		 String localVarPath = "/apis/extensions/v1beta1/namespaces/{namespace}/deployments"
		            .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));
		for(PredictorDef p : mlDep.getSpec().getPredictorsList())
		{
			String serviceLabel = getKubernetesDeploymentId(mlDep.getSpec().getName(),p.getName(), false);
			Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
				.setMetadata(ObjectMeta.newBuilder().setName("dep").putLabels(MLDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel))
				.setSpec(DeploymentSpec.newBuilder().setTemplate(p.getComponentSpec()).setReplicas(1)).build();
			
			
			
			
			logger.info(ProtoBufUtils.toJson(deployment));
			
			ObjectOrStatus os = pc.create(deployment, localVarPath, "extensions/v1beta1", "Deployment");
			String statusJson = ProtoBufUtils.toJson(os.status);
			logger.info(statusJson);
		}
	}
	
	public static ExtensionsV1beta1Deployment convertProtoToModel(Deployment protoDeployment) throws InvalidProtocolBufferException, MLDeploymentException
	{
		 Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
		 String ptsJson = jsonPrinter.print(protoDeployment);
		 JSON json = new JSON();
		 Type returnType = new TypeToken<ExtensionsV1beta1Deployment>(){}.getType();
		 ExtensionsV1beta1Deployment depSpec = (ExtensionsV1beta1Deployment) json.deserialize(ptsJson, returnType);
		 //return fixProbes(protoTemplateSpec, podTemplateSpec);
		 return depSpec;
	}
	
	@Test @Ignore
	public void convertTest() throws IOException, MLDeploymentException
	{
		MLDeploymentOperator op = new MLDeploymentOperatorImpl(getProps());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		MLDeployment mlDep = MLDeploymentUtils.jsonToMLDeployment(jsonStr);
		mlDep = op.defaulting(mlDep);
		for(PredictorDef p : mlDep.getSpec().getPredictorsList())
		{
			String serviceLabel = getKubernetesDeploymentId(mlDep.getSpec().getName(),p.getName(), false);
			Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
				.setMetadata(ObjectMeta.newBuilder().setName("dep").putLabels(MLDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel))
				.setSpec(DeploymentSpec.newBuilder().setTemplate(p.getComponentSpec()).setReplicas(1)).build();

			logger.info(ProtoBufUtils.toJson(deployment));
			
			ExtensionsV1beta1Deployment depSpec = convertProtoToModel(deployment);
			logger.info(depSpec.toString());
		}
	}
	
	
}
