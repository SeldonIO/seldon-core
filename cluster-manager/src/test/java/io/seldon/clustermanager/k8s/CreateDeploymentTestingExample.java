package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.lang.reflect.Type;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.gson.reflect.TypeToken;
import com.google.protobuf.util.JsonFormat;
import com.google.protobuf.util.JsonFormat.Printer;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.JSON;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.kubernetes.client.models.ExtensionsV1beta1DeploymentSpec;
import io.kubernetes.client.models.V1LabelSelector;
import io.kubernetes.client.models.V1ObjectMeta;
import io.kubernetes.client.models.V1PodTemplateSpec;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.kubernetes.client.util.Config;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public class CreateDeploymentTestingExample {
	private final static Logger logger = LoggerFactory.getLogger(CreateDeploymentTestingExample.class);
	
	 static String readFile(String path, Charset encoding) 
			  throws IOException 
	 {
		 byte[] encoded = Files.readAllBytes(Paths.get(path));
		 return new String(encoded, encoding);
	 }	

	 public static void main(String[] args) throws IOException, ApiException {
	        
		 ApiClient client = Config.defaultClient();
		 ExtensionsV1beta1Api api = new ExtensionsV1beta1Api(client);
		 
		 String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		 MLDeployment mldep = MLDeploymentUtils.jsonToMLDeployment(jsonStr);
		 PodTemplateSpec protoTemplateSpec = mldep.getSpec().getPredictors(0).getComponentSpec();
		 
		 Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
		 String ptsJson = jsonPrinter.print(protoTemplateSpec);
		 logger.info(ptsJson);

		 JSON json = new JSON();
		 Type returnType = new TypeToken<V1PodTemplateSpec>(){}.getType();
		 V1PodTemplateSpec podTemplate = (V1PodTemplateSpec) json.deserialize(ptsJson, returnType);
		 
		 logger.info("Image is "+podTemplate.getSpec().getContainers().get(0).getImage());

		 ExtensionsV1beta1DeploymentSpec depSpec = new ExtensionsV1beta1DeploymentSpec().template(podTemplate).replicas(2).selector(new V1LabelSelector().putMatchLabelsItem("app", "seldon"));
		 Map<String,String> labels = new HashMap<>();
		 ExtensionsV1beta1Deployment dep = new ExtensionsV1beta1Deployment().apiVersion("extensions/v1beta1").kind("Deployment")
				 	.metadata(new V1ObjectMeta().name("mydep").putLabelsItem("app", "myapp"))
				 	.spec(depSpec);
		 
		 try
		 {
			 ExtensionsV1beta1Deployment depRead = api.readNamespacedDeployment("mydep", "default", null, null, null);
			 ExtensionsV1beta1DeploymentList l =  api.listNamespacedDeployment("default", null, null, null, false, "app=myapp", 1, null, null, false);
			 if (l.getItems().size() == 0)
			 {
				 logger.info("Creating");
				 api.createNamespacedDeployment("default", dep, null);
			 }
			 else
			 {
				 logger.info("replacing");
				 api.replaceNamespacedDeployment("mydep", "default", dep, null);
			 }
		 }
		 catch (ApiException e)
		 {
			 logger.error("Error",e);
		 }

		 //
		 
	
		 
		 
	    }
}
