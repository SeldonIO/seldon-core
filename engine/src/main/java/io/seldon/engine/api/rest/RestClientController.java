package io.seldon.engine.api.rest;

import java.util.concurrent.ExecutionException;

import javax.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.engine.exception.APIException;
import io.seldon.engine.exception.APIException.ApiExceptionType;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.service.PredictionService;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Request;
import io.seldon.protos.PredictionProtos.Response;

@RestController
public class RestClientController {
	
	private static Logger logger = LoggerFactory.getLogger(RestClientController.class.getName());
	
	@Autowired
	private PredictionService predictionService;
	
	private boolean ready = false;
	
	 @PostConstruct
	 public void init(){
		 ready = true;
	 }	
	
	@RequestMapping("/")
    String home() {
        return "Hello World!!";
    }
	
	@RequestMapping(value = "/ping", method = RequestMethod.GET)
    String ping() {
        return "pong";
    }
	
	@RequestMapping("/ready")
	ResponseEntity<String> ready() {
		
		HttpHeaders responseHeaders = new HttpHeaders();
		HttpStatus httpStatus;
		String ret;
		if (ready)
		{
			httpStatus = HttpStatus.OK;
			ret = "ready";
		}
		else
		{
			httpStatus = HttpStatus.SERVICE_UNAVAILABLE;
			ret = "Service unavailable";
		}
		ResponseEntity<String> responseEntity = new ResponseEntity<String>(ret, responseHeaders, httpStatus);
		return responseEntity;
    }
	
	@RequestMapping("/pause")
    String pause() {	    
		ready = false;
        return "paused";
    }
	
	@RequestMapping("/unpause")
    String unpause() {	    
		ready = true;
        return "unpaused";
    }

	
	@RequestMapping(value = "/api/v0.1/predictions", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
    public ResponseEntity<String> predictions(RequestEntity<String> requestEntity)
	{
		Request request;
		try
		{
			Request.Builder builder = Request.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, requestEntity.getBody() );
			request = builder.build();
		} 
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad request",e);
			throw new APIException(ApiExceptionType.ENGINE_INVALID_JSON,requestEntity.getBody());
		}

		try
		{
			Response response = predictionService.predict(request);
			String json = ProtoBufUtils.toJson(response);
			return new ResponseEntity<String>(json,HttpStatus.OK);
		}
		 catch (InterruptedException e) {
			throw new APIException(ApiExceptionType.ENGINE_INTERRUPTED,e.getMessage());
		} catch (ExecutionException e) {
			if (e.getCause().getClass() == APIException.class){
				throw (APIException) e.getCause();
			}
			else
			{
				throw new APIException(ApiExceptionType.ENGINE_EXECUTION_FAILURE,e.getMessage());
			}
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
	
	@RequestMapping(value= "/api/v0.1/feedback", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
	public ResponseEntity<String>  feedback(RequestEntity<String> requestEntity) {
		Feedback feedback;
		
		try
		{
			Feedback.Builder builder = Feedback.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, requestEntity.getBody() );
			feedback = builder.build();
		} 
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad request",e);
			throw new APIException(ApiExceptionType.ENGINE_INVALID_JSON,requestEntity.getBody());
		}
		
		try
		{
			predictionService.sendFeedback(feedback);
			String json = "{}";
			return new ResponseEntity<String>(json,HttpStatus.OK);
		}
		 catch (InterruptedException e) {
			throw new APIException(ApiExceptionType.ENGINE_INTERRUPTED,e.getMessage());
		} catch (ExecutionException e) {
			if (e.getCause().getClass() == APIException.class){
				throw (APIException) e.getCause();
			}
			else
			{
				throw new APIException(ApiExceptionType.ENGINE_EXECUTION_FAILURE,e.getMessage());
			}
		}
    }
	
	@RequestMapping("/api/v0.1/events")
    String events() {
        return "Hello World!";
    }
}
