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
package io.seldon.apife.api.rest;

import static java.util.Arrays.asList;
import java.security.Principal;
import java.util.concurrent.ThreadLocalRandom;

import javax.annotation.PostConstruct;

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

import com.google.protobuf.InvalidProtocolBufferException;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Metrics;
import io.micrometer.core.instrument.Tag;
import io.seldon.apife.exception.SeldonAPIException;
import io.seldon.apife.exception.SeldonAPIException.ApiExceptionType;
import io.seldon.apife.kafka.KafkaRequestResponseProducer;
import io.seldon.apife.metrics.AuthorizedWebMvcTagsProvider;
import io.seldon.apife.pb.ProtoBufUtils;
import io.seldon.apife.service.PredictionService;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.RequestResponse;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@RestController
public class RestClientController {
	
	private final Logger logger = LoggerFactory.getLogger(this.getClass());
	
	@Autowired
	private PredictionService predictionService;
	
	@Autowired
	private KafkaRequestResponseProducer kafkaProducer;
	
	@Autowired
	AuthorizedWebMvcTagsProvider tagsProvider;
	
	private boolean ready = false;
	
	 @PostConstruct
	 public void init(){
		 ready = true;
	 }	
	

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
	
	@RequestMapping("/ping")
    String ping() {	    
        return "pong";
    }
	
	
	@RequestMapping(value = "/api/v0.1/predictions", method = RequestMethod.POST, consumes = "application/json; charset=utf-8", produces = "application/json; charset=utf-8")
	    public ResponseEntity<String> prediction(RequestEntity<String> requestEntity,Principal principal) {
		
		String clientId = principal.getName();
		String json = requestEntity.getBody();
		logger.info(String.format("[%s] [%s] [%s] [%s]", "POST", requestEntity.getUrl().getPath(), clientId, json));
		
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
			throw new SeldonAPIException(ApiExceptionType.APIFE_INVALID_JSON,requestEntity.getBody());
		}
		
		HttpStatus httpStatus = HttpStatus.OK;
		
		// At present passes JSON string. Could use gRPC?
		String ret = predictionService.predict(json,clientId);
		
		SeldonMessage response;
		try
		{
			SeldonMessage.Builder builder = SeldonMessage.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, ret);
			response = builder.build();
		}
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad response",e);
			throw new SeldonAPIException(ApiExceptionType.APIFE_INVALID_RESPONSE_JSON,requestEntity.getBody());
		}
		
		kafkaProducer.send(clientId,RequestResponse.newBuilder().setRequest(request).setResponse(response).build());
		
		
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
		Feedback feedback;
		try
		{
			Feedback.Builder builder = Feedback.newBuilder();
			ProtoBufUtils.updateMessageBuilderFromJson(builder, requestEntity.getBody() );
			feedback = builder.build();
			Iterable<Tag> tags = asList(tagsProvider.principal(clientId),tagsProvider.deploymentName(clientId));
			Counter.builder("seldon_api_ingress_server_feedback_reward").tags(tags).register(Metrics.globalRegistry).increment(feedback.getReward());
			Counter.builder("seldon_api_ingress_server_feedback").tags(tags).register(Metrics.globalRegistry).increment();
		} 
		catch (InvalidProtocolBufferException e) 
		{
			logger.error("Bad request",e);
			throw new SeldonAPIException(ApiExceptionType.APIFE_INVALID_RESPONSE_JSON,requestEntity.getBody());
		}
		
		predictionService.sendFeedback(json,clientId);
    }
	
}
