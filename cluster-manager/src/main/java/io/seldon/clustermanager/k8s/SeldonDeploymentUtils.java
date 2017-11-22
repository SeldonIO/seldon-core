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

import io.kubernetes.client.JSON;
import io.kubernetes.client.models.V1PodTemplateSpec;
import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Resource.Quantity;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.clustermanager.pb.IntOrStringUtils;
import io.seldon.clustermanager.pb.JsonFormat;
import io.seldon.clustermanager.pb.JsonFormat.Printer;
import io.seldon.clustermanager.pb.QuantityUtils;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class SeldonDeploymentUtils {
	private  static Logger logger = LoggerFactory.getLogger(SeldonDeploymentUtils.class.getName());
	
	private static <T> T convertProtoToModel(Message m,Type type) throws InvalidProtocolBufferException
	{
		Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
		 String ptsJson = jsonPrinter.print(m);
		 JSON json = new JSON();
		 return (T) json.deserialize(ptsJson, type);
	}
	
	
	
	public static V1PodTemplateSpec convertProtoToModel(PodTemplateSpec protoTemplateSpec) throws InvalidProtocolBufferException, SeldonDeploymentException
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
	
	public static SeldonDeployment jsonToMLDeployment(String json) throws InvalidProtocolBufferException {
		String jsonModified = removeCreationTimestampField(json);
		SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder();
		JsonFormat.parser().ignoringUnknownFields()
			.usingTypeParser(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringParser())
			.usingTypeParser(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityParser())
			.merge(jsonModified, mlBuilder);
		return mlBuilder.build();
	}
	
	public static String toJson(SeldonDeployment mlDep) throws InvalidProtocolBufferException
	{
		Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames()
				.usingTypeConverter(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringConverter())
				.usingTypeConverter(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityConverter());
		return jsonPrinter.print(mlDep);
				
	}
	
}
