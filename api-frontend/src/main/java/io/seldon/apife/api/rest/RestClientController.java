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
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import com.codahale.metrics.annotation.Timed;
import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.apife.exception.APIException;
import io.seldon.apife.exception.APIException.ApiExceptionType;
import io.seldon.apife.kafka.KafkaRequestResponseProducer;
import io.seldon.apife.pb.ProtoBufUtils;
import io.seldon.apife.service.PredictionService;
import io.seldon.protos.PredictionProtos.PredictionFeedbackDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionRequestResponseDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionFeedbackDef;

@RestController
public class RestClientController {
	
	private final Logger logger = LoggerFactory.getLogger(this.getClass());
	
	@Autowired
	private PredictionService predictionService;
	
	@Autowired
	private KafkaRequestResponseProducer kafkaProducer;
	
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
	
	
	@Timed
	@RequestMapping(value = "/api/v0.1/predictions", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
	    public ResponseEntity<String> prediction(RequestEntity<String> requestEntity,Principal principal) {
		
		String clientId = principal.getName();
		String json = requestEntity.getBody();
		logger.info(String.format("[%s] [%s] [%s] [%s]", "POST", requestEntity.getUrl().getPath(), clientId, json));
		
		PredictionRequestDef request;
		try
		{
			PredictionRequestDef.Builder builder = PredictionRequestDef.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, requestEntity.getBody() );
			request = builder.build();
		} 
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad request",e);
			throw new APIException(ApiExceptionType.APIFE_INVALID_JSON,requestEntity.getBody());
		}
		
		HttpStatus httpStatus = HttpStatus.OK;
		
		// At present passes JSON string. Could use gRPC?
		String ret = predictionService.predict(json,clientId);
		
		PredictionResponseDef response;
		try
		{
			PredictionResponseDef.Builder builder = PredictionResponseDef.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, ret);
			response = builder.build();
		}
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad response",e);
			throw new APIException(ApiExceptionType.APIFE_INVALID_RESPONSE_JSON,requestEntity.getBody());
		}
		
		kafkaProducer.send(clientId,PredictionRequestResponseDef.newBuilder().setRequest(request).setResponse(response).build());
		
		
		HttpHeaders responseHeaders = new HttpHeaders();
		responseHeaders.setContentType(MediaType.APPLICATION_JSON);
		ResponseEntity<String> responseEntity = new ResponseEntity<String>(ret, responseHeaders, httpStatus);

		return responseEntity;

	}

	
	@RequestMapping(value = "/api/v0.1/feedback", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
	@ResponseStatus(value = HttpStatus.OK)
	public void feedback(RequestEntity<String> requestEntity, Principal principal) 
	{
		String clientId = principal.getName();
		String json = requestEntity.getBody();
		PredictionFeedbackDef feedback;
		try
		{
			PredictionFeedbackDef.Builder builder = PredictionFeedbackDef.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, requestEntity.getBody() );
			feedback = builder.build();
		} 
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad request",e);
			throw new APIException(ApiExceptionType.APIFE_INVALID_RESPONSE_JSON,requestEntity.getBody());
		}
		
		predictionService.sendFeedback(json,clientId);
    }
	
}
