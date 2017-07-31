package io.seldon.engine.api.rest;

import java.io.IOException;
import java.util.concurrent.ExecutionException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParseException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.common.primitives.Doubles;
import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.engine.exception.APIException;
import io.seldon.engine.exception.APIException.ApiExceptionType;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.service.PredictionService;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;

@RestController
public class RestClientController {
	
	private static Logger logger = LoggerFactory.getLogger(RestClientController.class.getName());
	
	@Autowired
	private PredictionService predictionService;
	
	@RequestMapping("/")
    String home() {
        return "Hello World!";
    }
	
	@RequestMapping("/ping")
    String ping() {
        return "pong";
    }
	
	private String parseAndConvertJSON(String json) 
	{
		try
		{
			ObjectMapper mapper = new ObjectMapper();
			JsonFactory factory = mapper.getFactory();
			JsonParser parser = factory.createParser(json);
			JsonNode j = mapper.readTree(parser);
			if (j.has("request") && j.get("request").has("values"))
			{
				JsonNode values = j.get("request").get("values");
				double[][] v = mapper.readValue(values.toString(),double[][].class);
				double[] vs = Doubles.concat(v);
				int[] shape = {v.length,v[0].length };
				((ObjectNode) j.get("request")).replace("values",mapper.valueToTree(vs));
				((ObjectNode) j.get("request")).set("shape",mapper.valueToTree(shape));
				return j.toString();
			}
			else
				return null;
		} catch (JsonParseException e) {
			return null;
		} catch (IOException e) {
			return null;
		}
	}
	
	
	@RequestMapping(value = "/api/v0.1/predictions", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
    public ResponseEntity<String> predictions(RequestEntity<String> requestEntity) 
	{
		PredictionRequestDef request;
		try
		{
			PredictionRequestDef.Builder builder = PredictionRequestDef.newBuilder();
			String validJson = parseAndConvertJSON(requestEntity.getBody());
			if (validJson == null)
				validJson = requestEntity.getBody();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, validJson );
			request = builder.build();
		} 
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad request",e);
			throw new APIException(ApiExceptionType.ENGINE_INVALID_JSON,requestEntity.getBody());
		}

		try
		{
			PredictionResponseDef response = predictionService.predict(request);
			String json = ProtoBufUtils.toJson(response);
			return new ResponseEntity<String>(json,HttpStatus.OK);
		}
		 catch (InterruptedException e) {
			throw new APIException(ApiExceptionType.ENGINE_INTERTUPTED,e.getMessage());
		} catch (ExecutionException e) {
			throw new APIException(ApiExceptionType.ENGINE_EXECUTION_FAILURE,e.getMessage());
		} catch (InvalidProtocolBufferException e) {
			throw new APIException(ApiExceptionType.ENGINE_INVALID_JSON,"");
		} 

	}
	
	
	/*
	@RequestMapping(value="/api/v0.1/predictions", method = RequestMethod.POST)
    public @ResponseBody
    PredictionServiceReturn predictions(@RequestBody PredictionServiceRequest request, HttpServletRequest req) throws InterruptedException, ExecutionException {

        //TODO: Check authentication here
		
		
		return predictionService.predict(request);
		
    }
    */
	
	@RequestMapping("/api/v0.1/feedback")
    String feedback() {
        return "Hello World!";
    }
	
	@RequestMapping("/api/v0.1/events")
    String events() {
        return "Hello World!";
    }
}
