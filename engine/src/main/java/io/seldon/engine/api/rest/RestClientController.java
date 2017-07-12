package io.seldon.engine.api.rest;

import java.io.IOException;
import java.util.concurrent.ExecutionException;

import javax.servlet.http.HttpServletRequest;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.ResponseBody;
import org.springframework.web.bind.annotation.RestController;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParseException;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.seldon.engine.exception.APIException;
import io.seldon.engine.service.PredictionService;
import io.seldon.engine.service.PredictionServiceRequest;
import io.seldon.engine.service.PredictionServiceReturn;

@RestController
public class RestClientController {
	
	@Autowired
	private PredictionService predictionService;
	
	
	private JsonNode getValidatedJson(String jsonRaw) 
	{
		try
		{
			ObjectMapper mapper = new ObjectMapper();
			JsonFactory factory = mapper.getFactory();
			JsonParser parser = factory.createParser(jsonRaw);
			JsonNode actualObj = mapper.readTree(parser);
			return actualObj;
		} catch (JsonParseException e) {
			throw new APIException(io.seldon.engine.exception.APIException.ApiExceptionType.APIFE_INVALID_JSON,jsonRaw);
		} catch (IOException e) {
			throw new APIException(APIException.ApiExceptionType.APIFE_INVALID_JSON,jsonRaw);
		}
	}
	
	
	@RequestMapping("/")
    String home() {
        return "Hello World!";
    }
	
	/*
	@RequestMapping(value = "/api/v0.1/predictions", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
    public ResponseEntity<String> test(RequestEntity<String> requestEntity,Principal principal) {
	}
	*/
	
	@RequestMapping(value="/api/v0.1/predictions", method = RequestMethod.POST)
    public @ResponseBody
    PredictionServiceReturn predictions(@RequestBody PredictionServiceRequest request, HttpServletRequest req) throws InterruptedException, ExecutionException {

        //TODO: Check authentication here
		
		
		return predictionService.predict(request);
		
    }
	
	@RequestMapping("/api/v0.1/feedback")
    String feedback() {
        return "Hello World!";
    }
	
	@RequestMapping("/api/v0.1/events")
    String events() {
        return "Hello World!";
    }
}
