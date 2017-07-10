package io.seldon.apife.api.rest;

import java.io.IOException;
import java.security.Principal;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
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

import io.seldon.apife.exception.APIException;
import io.seldon.apife.service.PredictionService;

@RestController
public class RestClientController {
	
	private final Logger logger = LoggerFactory.getLogger(this.getClass());
	
	@Autowired
	private PredictionService predictionService;
	
	@RequestMapping("/")
    String home() {
        return "Hello World!";
    }
	
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
			throw new APIException(APIException.INVALID_JSON);
		} catch (IOException e) {
			throw new APIException(APIException.INVALID_JSON);
		}
	}
	
	
	@RequestMapping(value = "/api/v0.1/predictions", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
	    public ResponseEntity<String> test(RequestEntity<String> requestEntity,Principal principal) {
		
		String clientId = principal.getName();
		String json = requestEntity.getBody();
		logger.info(String.format("[%s] [%s] [%s] [%s]", "POST", requestEntity.getUrl().getPath(), clientId, json));
		
		HttpStatus httpStatus = HttpStatus.OK;
		
		String ret = predictionService.predict(json,clientId);
		
		HttpHeaders responseHeaders = new HttpHeaders();
		responseHeaders.setContentType(MediaType.APPLICATION_JSON);
		ResponseEntity<String> responseEntity = new ResponseEntity<String>(ret, responseHeaders, httpStatus);

		return responseEntity;

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
