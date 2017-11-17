package io.seldon.clustermanager.k8s;

import java.io.IOException;
import java.lang.reflect.Type;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParseException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.gson.reflect.TypeToken;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.util.JsonFormat;
import com.google.protobuf.util.JsonFormat.Printer;

import io.kubernetes.client.JSON;
import io.kubernetes.client.models.V1Container;
import io.kubernetes.client.models.V1HTTPGetAction;
import io.kubernetes.client.models.V1PodTemplateSpec;
import io.kubernetes.client.models.V1Probe;
import io.kubernetes.client.models.V1TCPSocketAction;
import io.kubernetes.client.proto.V1.HTTPGetAction;
import io.kubernetes.client.proto.V1.Handler;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.kubernetes.client.proto.V1.TCPSocketAction;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public class MLDeploymentUtils {
	private  static Logger logger = LoggerFactory.getLogger(MLDeploymentUtils.class.getName());
	
	private static <T> T convertProtoToModel(Message m,Type type) throws InvalidProtocolBufferException
	{
		Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
		 String ptsJson = jsonPrinter.print(m);
		 JSON json = new JSON();
		 return (T) json.deserialize(ptsJson, type);
	}
	
	/*
	private static void updateProbe(V1Probe probe,Handler handler,int cidx) throws InvalidProtocolBufferException, MLDeploymentException
	{
		if (handler.hasHttpGet())
		{
			Type returnType = new TypeToken<V1HTTPGetAction>(){}.getType();
			V1HTTPGetAction httpGet = convertProtoToModel(HTTPGetAction.newBuilder(handler.getHttpGet()).clearPort().build(), returnType);
			HTTPGetAction protoHttpGet = handler.getHttpGet();	
			if (protoHttpGet.hasPort())
			{
				if (protoHttpGet.getPort().hasStrVal())
					httpGet.setPort(protoHttpGet.getPort().getStrVal());
				else
					throw new MLDeploymentException("Container "+cidx+" has integer http port liveness probe which is not supported by kubernetes Java client");
			}
			probe.setHttpGet(httpGet);
		}
		else if (handler.hasTcpSocket())
		{
			Type returnType = new TypeToken<V1TCPSocketAction>(){}.getType();
			V1TCPSocketAction tcpSocket = convertProtoToModel(TCPSocketAction.newBuilder(handler.getTcpSocket()).clearPort().build(), returnType);
			TCPSocketAction protoTcpGet = handler.getTcpSocket();	
			if (protoTcpGet.hasPort())
			{
				if (protoTcpGet.getPort().hasStrVal())
					tcpSocket.setPort(protoTcpGet.getPort().getStrVal());
				else
					throw new MLDeploymentException("Container "+cidx+" has integer tcp port liveness probe which is not supported by kubernetes Java client");
			}
			probe.setTcpSocket(tcpSocket);
		}
	}

	
	private static V1PodTemplateSpec fixProbes(PodTemplateSpec protoTemplateSpec,V1PodTemplateSpec spec) throws MLDeploymentException, InvalidProtocolBufferException
	{
		int cidx = 0;
		for (V1Container c : spec.getSpec().getContainers())
		{
			if (c.getLivenessProbe() != null)
			{
				V1Probe probe = c.getLivenessProbe();
				Handler handler = protoTemplateSpec.getSpec().getContainers(cidx).getLivenessProbe().getHandler();
				updateProbe(probe, handler, cidx);
			}
			if (c.getReadinessProbe() != null)
			{
				V1Probe probe = c.getReadinessProbe();
				Handler handler = protoTemplateSpec.getSpec().getContainers(cidx).getReadinessProbe().getHandler();
				updateProbe(probe, handler, cidx);				
			}
		}
		return spec;
	}
		*/
	
	public static V1PodTemplateSpec convertProtoToModel(PodTemplateSpec protoTemplateSpec) throws InvalidProtocolBufferException, MLDeploymentException
	{
		 Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
		 String ptsJson = jsonPrinter.print(protoTemplateSpec);
		 JSON json = new JSON();
		 Type returnType = new TypeToken<V1PodTemplateSpec>(){}.getType();
		 V1PodTemplateSpec podTemplateSpec = (V1PodTemplateSpec) json.deserialize(ptsJson, returnType);
		 //return fixProbes(protoTemplateSpec, podTemplateSpec);
		 return podTemplateSpec;
	}

	
	private static String removeCreationTimestampField(String json)
	{
		try
		{
		ObjectMapper mapper = new ObjectMapper();
	    JsonFactory factory = mapper.getFactory();
	    JsonParser parser = factory.createParser(json);
	    JsonNode obj = mapper.readTree(parser);
	    if (obj.has("metadata") && obj.get("metadata").has("creationTimestamp"))
	    {
	    	((ObjectNode) obj.get("metadata")).remove("creationTimestamp");
	    	return mapper.writeValueAsString(obj);
	    }
	    else
	    	return json;
		} catch (JsonParseException e) {
			logger.error("Failed to remove creationTimestamp");
			return json;
		} catch (IOException e) {
			logger.error("Failed to remove creationTimestamp");
			return json;
		}
		
	}
	
	public static MLDeployment jsonToMLDeployment(String json) throws InvalidProtocolBufferException {
		String jsonModified = removeCreationTimestampField(json);
		MLDeployment.Builder mlBuilder = MLDeployment.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(mlBuilder, jsonModified);
		return mlBuilder.build();
	}
	
}
