package io.seldon.clustermanager.k8s;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParseException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public class MLDeploymentUtils {
	private  static Logger logger = LoggerFactory.getLogger(MLDeploymentUtils.class.getName());
	
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
