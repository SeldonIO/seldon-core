package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.io.InputStream;
import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;

import org.junit.Ignore;
import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.common.io.ByteStreams;
import com.google.common.primitives.Bytes;
import com.google.gson.reflect.TypeToken;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.util.JsonFormat;
import com.google.protobuf.util.JsonFormat.Printer;
import com.squareup.okhttp.MediaType;
import com.squareup.okhttp.Request;
import com.squareup.okhttp.RequestBody;
import com.squareup.okhttp.Response;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.JSON;
import io.kubernetes.client.Pair;
import io.kubernetes.client.ProtoClient;
import io.kubernetes.client.ProtoClient.ObjectOrStatus;
import io.kubernetes.client.apis.ExtensionsV1beta1Api;
import io.kubernetes.client.models.ExtensionsV1beta1Deployment;
import io.kubernetes.client.models.V1DeleteOptions;
import io.kubernetes.client.models.V1Status;
import io.kubernetes.client.proto.Meta.DeleteOptions;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.kubernetes.client.proto.Meta.Status;
import io.kubernetes.client.proto.Runtime.TypeMeta;
import io.kubernetes.client.proto.Runtime.Unknown;
import io.kubernetes.client.proto.V1.Service;
import io.kubernetes.client.proto.V1.ServicePort;
import io.kubernetes.client.proto.V1.ServiceSpec;
import io.kubernetes.client.proto.V1beta1Extensions;
import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.kubernetes.client.proto.V1beta1Extensions.DeploymentSpec;
import io.kubernetes.client.util.Config;
import io.seldon.clustermanager.AppTest;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;


public class ProtoClientTest extends AppTest {

	private final static Logger logger = LoggerFactory.getLogger(ProtoClientTest.class);
	
	private static String getKubernetesSeldonDeploymentId(String deploymentName, boolean isCanary) {
        return "sd-" + deploymentName + "-" + ((isCanary) ? "c" : "p");
    }
	
	private static String getKubernetesDeploymentId(String deploymentName,String predictorName, boolean isCanary) {
		return "sd-" + deploymentName + "-" + predictorName + "-" + ((isCanary) ? "c" : "p");
	}
	
	@Test 
	public void protoTest() throws IOException, ApiException
	{
		ApiClient apiClient = Config.defaultClient();
		ProtoClient pc = new ProtoClient(apiClient);
		
		SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getProps());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		mlDep = op.defaulting(mlDep);
		String localVarPath = "/apis/extensions/v1beta1/namespaces/{namespace}/deployments"
		            .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));
		String serviceLabel = getKubernetesSeldonDeploymentId(mlDep.getSpec().getName(), false);
		for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
		{


			Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
				.setMetadata(ObjectMeta.newBuilder().setName("dep").putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel))
				.setSpec(DeploymentSpec.newBuilder().setTemplate(p.getComponentSpec()).setReplicas(1)).build();
			
			// create path and map variables
	        String localVarPath2 = "/apis/extensions/v1beta1/namespaces/{namespace}/deployments/{name}"
	            .replaceAll("\\{" + "name" + "\\}", apiClient.escapeString("dep"))
	            .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));

			
			ObjectOrStatus os = pc.list(Deployment.newBuilder(),localVarPath2);		
			if (os.status != null)
			{
				if (os.status.getCode() == 404)
				{
					os = pc.create(deployment, localVarPath, "extensions/v1beta1", "Deployment");
					if (os.status != null)
					{
						logger.info("Possible error creating deployment "+ProtoBufUtils.toJson(os.status));
					}
					else
					{
						logger.info("Created deployment:"+ProtoBufUtils.toJson(os.object));
					}					
				}
				else
					logger.info("Error listing deployment:"+ProtoBufUtils.toJson(os.status));
			}
			else
			{
				logger.info("Returned object:"+ProtoBufUtils.toJson(os.object));
				Deployment.Builder b = Deployment.newBuilder().mergeFrom(os.object);
				if (b.getSpec().getReplicas() == 2)
				{
					logger.info("delete resource");
					if (true)
					{
						DeleteOptions deleteOptions = DeleteOptions.newBuilder().setPropagationPolicy("Foreground").build();
						String localVarPath4 = "/apis/extensions/v1beta1/namespaces/{namespace}/deployments/{name}"
					            .replaceAll("\\{" + "name" + "\\}", apiClient.escapeString("dep"))
					            .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));
						//Status status = pc.delete(Deployment.newBuilder(),localVarPath4);
						os = pc.delete(Deployment.newBuilder(),localVarPath4,deleteOptions);
						//os = pc.delete(Deployment.newBuilder(),localVarPath4,null);
						//os = request(apiClient,Deployment.newBuilder(),localVarPath4,"DELETE",deleteOptions, "v1", "DeleteOptions");
						if (os.status != null)
							logger.info("Error during delete  "+ProtoBufUtils.toJson(os.status));
						else
						{
							logger.info("deleted deployment:"+ProtoBufUtils.toJson(os.object));
						}
					}
					else
					{
						ExtensionsV1beta1Api api = new ExtensionsV1beta1Api(apiClient);
						
						V1Status status = api.deleteNamespacedDeployment("dep","default",new V1DeleteOptions().propagationPolicy("Foreground"),null,null,null,null);
						logger.info(status.toString());
					}
				}
				else
				{
					b.getSpecBuilder().setReplicas(2);
					Deployment d2 = b.build();
					logger.info(ProtoBufUtils.toJson(d2));
					//update deployment
					//deployment = Deployment.newBuilder(deployment).setSpec(DeploymentSpec.newBuilder(deployment.getSpec()).setReplicas(2).build()).build();
					int reps = deployment.getSpec().getReplicas();
					String localVarPath4 = "/apis/extensions/v1beta1/namespaces/{namespace}/deployments/{name}"
				            .replaceAll("\\{" + "name" + "\\}", apiClient.escapeString("dep"))
				            .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));
					String localVarPath3 = "/apis/extensions/v1beta1/namespaces/{namespace}/deployments/{name}/status"
				            .replaceAll("\\{" + "name" + "\\}", apiClient.escapeString("dep"))
				            .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));
					os = pc.update(d2,localVarPath4, "extensions/v1beta1", "Deployment");
					//os = pc.request(d2.newBuilderForType(),localVarPath4,"PUT",d2, "extensions/v1beta1", "Deployment");
					if (os.status != null)
					{
						logger.info("Possible error updating deployment "+ProtoBufUtils.toJson(os.status));
					}
					else
					{
						logger.info("Updated deployment:"+ProtoBufUtils.toJson(os.object));
					}
				}
				
			}
		}
		
		Service s = Service.newBuilder()
                .setMetadata(ObjectMeta.newBuilder()
                        .setName(mlDep.getSpec().getName())
                        .putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel)
                        .putLabels("seldon-deployment-id", mlDep.getSpec().getName())
                        )
                .setSpec(ServiceSpec.newBuilder()
                        .addPorts(ServicePort.newBuilder()
                                .setProtocol("TCP")
                                .setPort(9000)
                                .setTargetPort(IntOrString.newBuilder().setIntVal(9000))
                                .setName("http")
                                )
                        .setType("ClusterIP")
                        .putSelector(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP,serviceLabel)
                        )
            .build();

		final String serviceApiPath = "/api/v1/namespaces/{namespace}/services/{name}"
                .replaceAll("\\{" + "name" + "\\}", apiClient.escapeString(s.getMetadata().getName()))
                .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));

            
		ObjectOrStatus os = pc.list(Service.newBuilder(),serviceApiPath);     
		if (os.status != null)
		{
		    if (os.status.getCode() == 404)
		    {
		        String serviceCreateApiPath = "/api/v1/namespaces/{namespace}/services"
		                .replaceAll("\\{" + "namespace" + "\\}", apiClient.escapeString("default"));
		        os = pc.create(s, serviceCreateApiPath, "v1", "Service");
		        if (os.status != null)
		        {
		            logger.info("Possible error creating service "+ProtoBufUtils.toJson(os.status));
		        }
		        else
		        {
		            logger.info("Created service:"+ProtoBufUtils.toJson(os.object));
		        }                   
		    }
		    else
		        logger.info("Error listing service:"+ProtoBufUtils.toJson(os.status));
		}
		
	}
	 // This isn't really documented anywhere except the code, but
    // the proto-buf format is:
    //   * 4 byte magic number
    //   * Protocol Buffer encoded object of type runtime.Unknown
    //   * the 'raw' field in that object contains a Protocol Buffer
    //     encoding of the actual object.
    // TODO: Document this somewhere proper.

    private byte[] encode(Message msg, String apiVersion, String kind) {
        // It is unfortunate that we have to include apiVersion and kind,
        // since we should be able to extract it from the Message, but
        // for now at least, those fields are missing from the proto-buffer.
        Unknown u = Unknown.newBuilder().setTypeMeta(TypeMeta.newBuilder().setApiVersion(apiVersion).setKind(kind))
                .setRaw(msg.toByteString()).build();
        return Bytes.concat(MAGIC, u.toByteArray());
    }
    
    private static final byte[] MAGIC = new byte[] { 0x6b, 0x38, 0x73, 0x00 };
    private static final String MEDIA_TYPE = "application/vnd.kubernetes.protobuf";


    private Unknown parse(InputStream stream) throws ApiException, IOException {
        byte[] magic = new byte[4];
        ByteStreams.readFully(stream, magic);
        if (!Arrays.equals(magic, MAGIC)) {
            throw new ApiException("Unexpected magic number: " + magic);
        }
        return Unknown.parseFrom(stream);
    }
	
	 public <T extends Message> ObjectOrStatus<T> request(ApiClient apiClient,T.Builder builder, String path, String method, T body, String apiVersion,
	            String kind) throws ApiException, IOException {
	        HashMap<String, String> headers = new HashMap<String, String>();
	        headers.put("Content-type", MEDIA_TYPE);
	        headers.put("Accept", MEDIA_TYPE);
	        Request request = apiClient.buildRequest(path, method, new ArrayList<Pair>(), new ArrayList<Pair>(), null,
	                headers, new HashMap<String, Object>(), new String[0], null);
	        if (body != null) {
	            byte[] bytes = encode(body, apiVersion, kind);
	            if (method.equals("POST"))
	            	request = request.newBuilder().post(RequestBody.create(MediaType.parse(MEDIA_TYPE), bytes)).build();
	            else if ("PUT".equals(method))
	            	request = request.newBuilder().put(RequestBody.create(MediaType.parse(MEDIA_TYPE), bytes)).build();
	            else if ("DELETE".equals(method))
	            	request = request.newBuilder().delete(RequestBody.create(MediaType.parse(MEDIA_TYPE), bytes)).build();
	        }
	        Response resp = apiClient.getHttpClient().newCall(request).execute();
	        Unknown u = parse(resp.body().byteStream());
	        resp.body().close();

	        if (u.getTypeMeta().getApiVersion().equals("v1") &&
	            u.getTypeMeta().getKind().equals("Status")) {
	            Status status = Status.newBuilder().mergeFrom(u.getRaw()).build();
	            return new ObjectOrStatus(null, status);
	        }

	        return new ObjectOrStatus((T) builder.mergeFrom(u.getRaw()).build(), null);
	    }
	
	public static ExtensionsV1beta1Deployment convertProtoToModel(Deployment protoDeployment) throws InvalidProtocolBufferException, SeldonDeploymentException
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
	public void convertTest() throws IOException, SeldonDeploymentException
	{
		SeldonDeploymentOperator op = new SeldonDeploymentOperatorImpl(getProps());
		String jsonStr = readFile("src/test/resources/mldeployment_1.json",StandardCharsets.UTF_8);
		SeldonDeployment mlDep = SeldonDeploymentUtils.jsonToSeldonDeployment(jsonStr);
		mlDep = op.defaulting(mlDep);
		for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
		{
			String serviceLabel = getKubernetesDeploymentId(mlDep.getSpec().getName(),p.getName(), false);
			Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
				.setMetadata(ObjectMeta.newBuilder().setName("dep").putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel))
				.setSpec(DeploymentSpec.newBuilder().setTemplate(p.getComponentSpec()).setReplicas(1)).build();

			logger.info(ProtoBufUtils.toJson(deployment));
			
			ExtensionsV1beta1Deployment depSpec = convertProtoToModel(deployment);
			logger.info(depSpec.toString());
		}
	}
	
	
}
