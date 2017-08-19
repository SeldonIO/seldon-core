package io.seldon.apife.api.rest;

import java.security.Principal;
import java.util.concurrent.ThreadLocalRandom;

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

import com.codahale.metrics.annotation.Timed;

import io.seldon.apife.service.PredictionService;

@RestController
public class RestClientController {
	
	private final Logger logger = LoggerFactory.getLogger(this.getClass());
	
	@Autowired
	private PredictionService predictionService;
	
	@Timed
	@RequestMapping("/")
    String home() {
	    
        int randomNum = ThreadLocalRandom.current().nextInt(0, 10);
        try {
            Thread.sleep(40 + randomNum); // simulate a delay 40~50 ms
        } catch (InterruptedException e) {
            e.printStackTrace();
        }

        return "Hello World!";
    }
	
	
	@RequestMapping("/ping")
    String ping() {	    
        return "pong";
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
        return "Not Implemented";
    }
	
}
