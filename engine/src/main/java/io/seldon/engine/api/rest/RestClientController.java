/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.api.rest;

import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;

import javax.annotation.PostConstruct;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

import com.google.protobuf.InvalidProtocolBufferException;

import io.micrometer.core.annotation.Timed;
import io.opentracing.Scope;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.exception.APIException.ApiExceptionType;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.service.PredictionService;
import io.seldon.engine.tracing.TracingProvider;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@RestController
public class RestClientController {
	
	private static Logger logger = LoggerFactory.getLogger(RestClientController.class.getName());


	@Autowired
	private PredictionService predictionService;
	
	@Autowired
	SeldonGraphReadyChecker readyChecker;
	
	@Autowired
	TracingProvider tracingProvider;
	
	private AtomicBoolean ready = new AtomicBoolean(false);
	
	 @PostConstruct
	 public void init(){
		 ready.set(true);
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
		if (ready.get() && readyChecker.getReady())
		{
			httpStatus = HttpStatus.OK;
			ret = "ready";
		}
		else
		{
			logger.warn("Not ready graph checker {}, controller {}",readyChecker.getReady(),ready.get());
			httpStatus = HttpStatus.SERVICE_UNAVAILABLE;
			ret = "Service unavailable";
		}
		ResponseEntity<String> responseEntity = new ResponseEntity<String>(ret, responseHeaders, httpStatus);
		return responseEntity;
    }

	@RequestMapping("/live")
	ResponseEntity<String> live() {

		HttpHeaders responseHeaders = new HttpHeaders();
		HttpStatus httpStatus;
		String ret  = "live";
		httpStatus = HttpStatus.OK;

		ResponseEntity<String> responseEntity = new ResponseEntity<String>(ret, responseHeaders, httpStatus);
		return responseEntity;
	}


	@RequestMapping("/pause")
    String pause() {	    
		ready.set(false);
        logger.warn("App Paused");
        return "paused";
    }
	
	@RequestMapping("/unpause")
    String unpause() {	    
		ready.set(true);
        logger.warn("App UnPaused");		
        return "unpaused";
    }

	@Timed
	@CrossOrigin(origins = "*")
	@RequestMapping(value = "/api/v0.1/predictions", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
    public ResponseEntity<String> predictions(RequestEntity<String> requestEntity)
	{
		logger.debug("Received predict request");
		Scope tracingScope = null;
		if (tracingProvider.isActive())
			tracingScope = tracingProvider.getTracer().buildSpan("/api/v0.1/predictions").startActive(true);
		try
		{
		SeldonMessage request;
		try
		{
			SeldonMessage.Builder builder = SeldonMessage.newBuilder();
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
			SeldonMessage response = predictionService.predict(request);
			String responseJson = ProtoBufUtils.toJson(response);
			return new ResponseEntity<String>(responseJson,HttpStatus.OK);
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
		finally
		{
			if (tracingScope != null)
				tracingScope.close();
		}

	}
	
	@Timed
	@CrossOrigin(origins = "*")
	@RequestMapping(value= "/api/v0.1/feedback", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
	public ResponseEntity<String>  feedback(RequestEntity<String> requestEntity) {
		logger.debug("Received feedback request");
		Scope tracingScope = null;
		if (tracingProvider.isActive())
			tracingScope = tracingProvider.getTracer().buildSpan("/api/v0.1/feedback").startActive(true);
		try
		{
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
		} catch (InvalidProtocolBufferException e) {
			throw new APIException(ApiExceptionType.ENGINE_INVALID_JSON,"");
		} 
		}
		finally
		{
			if (tracingScope != null)
				tracingScope.close();
		}

    }

}
