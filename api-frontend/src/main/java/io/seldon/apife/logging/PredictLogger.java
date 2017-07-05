package io.seldon.apife.logging;

import org.apache.log4j.Logger;
import org.springframework.stereotype.Component;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import io.seldon.apife.predictors.PredictorRequest;
import io.seldon.apife.service.PredictionServiceReturn;

@Component
public class PredictLogger {

	private static Logger predictLogger = Logger.getLogger( "PredictLogger" );
	
	
	public void log(String deployment,PredictorRequest request,PredictionServiceReturn response)
	{
		ObjectMapper mapper = new ObjectMapper();
		JsonNode prediction = mapper.valueToTree(response);
		JsonNode input = mapper.valueToTree(request);
		ObjectNode topNode = mapper.createObjectNode();
		topNode.put("deployment", deployment);
		topNode.put("input", input);
		topNode.put("prediction", prediction);
		predictLogger.info(topNode.toString());
	}
	
}
