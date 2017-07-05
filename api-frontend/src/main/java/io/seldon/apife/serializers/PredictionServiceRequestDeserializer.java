package io.seldon.apife.serializers;

import java.io.IOException;

import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.DeserializationContext;
import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.deser.std.StdDeserializer;
import com.fasterxml.jackson.databind.node.ObjectNode;

import io.seldon.apife.predictors.PredictorRequestJSON;
import io.seldon.apife.service.PredictionServiceRequest;
import io.seldon.apife.service.PredictionServiceRequestMeta;

public class PredictionServiceRequestDeserializer extends StdDeserializer<PredictionServiceRequest> {
	
	/**
	 * 
	 */
	private static final long serialVersionUID = 1L;

	public PredictionServiceRequestDeserializer() {
        super(PredictionServiceRequest.class);
    }

	@Override
	public PredictionServiceRequest deserialize(JsonParser jp, DeserializationContext ctxt)
			throws IOException, JsonProcessingException {
		// TODO Auto-generated method stub
		ObjectMapper mapper = (ObjectMapper) jp.getCodec();
        ObjectNode root = mapper.readTree(jp);
        
        PredictionServiceRequestMeta meta = mapper.readValue(root.get("meta").toString(),PredictionServiceRequestMeta.class);
        
        PredictorRequestJSON request;
        
        JsonNode requestNode = root.get("request");
        
        try {
        	request = mapper.readValue(requestNode.toString(), PredictorRequestJSON.class);
        	request.isDefault = false;
		}
		catch (JsonMappingException e) {
			request = new PredictorRequestJSON(requestNode.get("data").toString());
			request.isDefault = true;
		}
        
        return new PredictionServiceRequest(meta,request);
	}
	
}
