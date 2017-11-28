package io.seldon.engine.service;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.concurrent.TimeUnit;

import org.apache.commons.lang.NotImplementedException;
import org.apache.http.client.utils.URIBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.util.JsonFormat;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.predictors.PredictiveUnitState;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.ModelGrpc;
import io.seldon.protos.ModelGrpc.ModelBlockingStub;
import io.seldon.protos.RouterGrpc;
import io.seldon.protos.RouterGrpc.RouterBlockingStub;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Message.DataOneofCase;
import io.seldon.protos.PredictionProtos.Message;

@Service
public class InternalPredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(InternalPredictionService.class.getName());

	public static final String MODEL_NAME_HEADER = "Seldon-model-name"; 
	public static final String MODEL_IMAGE_HEADER = "Seldon-model-image"; 
	public static final String MODEL_VERSION_HEADER = "Seldon-model-version"; 
	
    ObjectMapper mapper = new ObjectMapper();
    
    RestTemplate restTemplate;
    
    
    
    @Autowired
    public InternalPredictionService(RestTemplate restTemplate){
    	this.restTemplate = restTemplate;
    	
    }
		
	public Message getPrediction(Message request, PredictiveUnitState state) throws JsonProcessingException, IOException{

		final Endpoint endpoint = state.endpoint;
		switch (endpoint.getType()){
			case REST:
				String dataString = ProtoBufUtils.toJson(request);
				boolean isDefault = false;
				if (request.getDataOneofCase() == DataOneofCase.DATA)
					isDefault = true;
				return getPredictionREST(dataString, state, endpoint, isDefault);
				
			case GRPC:
				return getPredictionGRPC(request,state,endpoint);
		}
		throw new APIException(APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,"no service available");
	}
	
	public Message getRouting(Message request, Endpoint endpoint){
		switch (endpoint.getType()){
			case REST:
				throw new NotImplementedException();
			case GRPC:
				return getRoutingGRPC(request, endpoint);
				
		}
		return null;
	}
	
	public void sendFeedback(Feedback feedback, Endpoint endpoint){
		switch (endpoint.getType()){
			case REST:
				throw new NotImplementedException();
			case GRPC:
				sendFeedbackGRPC(feedback, endpoint);
		}
		return;
	}
	
	public void sendFeedbackRouter(Feedback feedback, Endpoint endpoint){
		switch (endpoint.getType()){
			case REST:
				throw new NotImplementedException();
			case GRPC:
				sendFeedbackRouterGRPC(feedback, endpoint);
		}
		return;
	}
	
	private void sendFeedbackGRPC(Feedback feedback, Endpoint endpoint){
		ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();
		ModelBlockingStub stub =  ModelGrpc.newBlockingStub(channel).withDeadlineAfter(5, TimeUnit.SECONDS);
		
		stub.sendFeedback(feedback);
		
		return;
	}
	
	private void sendFeedbackRouterGRPC(Feedback feedback, Endpoint endpoint){
		ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();
		RouterBlockingStub stub =  RouterGrpc.newBlockingStub(channel).withDeadlineAfter(5, TimeUnit.SECONDS);
		
		stub.sendFeedback(feedback);
		
		return;
	}
	
	private Message getRoutingGRPC(Message request, Endpoint endpoint){
		ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();

		RouterBlockingStub stub =  RouterGrpc.newBlockingStub(channel).withDeadlineAfter(5, TimeUnit.SECONDS);
		Message routing;
		try {
			routing = stub.route(request);
		} catch (StatusRuntimeException e) 
		{
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ENDPOINT_URL,"Host: "+endpoint.getServiceHost()+" port:"+endpoint.getServicePort());
		}
		
		return routing;
	}
	
	public Message getPredictionGRPC(Message request, PredictiveUnitState state, Endpoint endpoint){
		ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();
		ModelBlockingStub stub =  ModelGrpc.newBlockingStub(channel).withDeadlineAfter(5, TimeUnit.SECONDS);
			
		Message response = stub.predict(request);
		return response;
	}
	
	
	
	public Message getPredictionREST(String dataString, PredictiveUnitState state, Endpoint endpoint, boolean isDefault){
		{
    		long timeNow = System.currentTimeMillis();
    		URI uri;
			try {
    			URIBuilder builder = new URIBuilder().setScheme("http")
    					.setHost(endpoint.getServiceHost())
    					.setPort(endpoint.getServicePort())
    					.setPath("/predict");

    			uri = builder.build();
    		} catch (URISyntaxException e) 
    		{
    			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ENDPOINT_URL,"Host: "+endpoint.getServiceHost()+" port:"+endpoint.getServicePort());
    		}
    		
    		try  
    		{
    			HttpHeaders headers = new HttpHeaders();
    			headers.setContentType(MediaType.APPLICATION_FORM_URLENCODED);
    			headers.add(MODEL_NAME_HEADER, state.name);
    			headers.add(MODEL_IMAGE_HEADER, state.imageName);
    			headers.add(MODEL_VERSION_HEADER, state.imageVersion);
    			
    			MultiValueMap<String, String> map= new LinkedMultiValueMap<String, String>();
    			map.add("json", dataString);
    			map.add("isDefault", Boolean.toString(isDefault));

    			HttpEntity<MultiValueMap<String, String>> request = new HttpEntity<MultiValueMap<String, String>>(map, headers);

    			logger.info("Requesting " + uri.toString());
    			ResponseEntity<String> httpResponse = restTemplate.postForEntity( uri, request , String.class );
    			
    			try
    			{
    				if(httpResponse.getStatusCode().is2xxSuccessful()) 
    				{
    				    Message.Builder builder = Message.newBuilder();
    				    String response = httpResponse.getBody();
    				    logger.info(response);
    				    JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
    				    return builder.build();
    				} 
    				else 
    				{
    					logger.error("Couldn't retrieve prediction from external prediction server -- bad http return code: " + httpResponse.getStatusCode());
    					throw new APIException(APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,String.format("Bad return code %d", httpResponse.getStatusCode()));
    				}
    			}
    			finally
    			{
    				if (logger.isDebugEnabled())
    					logger.debug("External prediction server took "+(System.currentTimeMillis()-timeNow) + "ms");
    			}
    		} 
    		catch (IOException e) 
    		{
    			logger.error("Couldn't retrieve prediction from external prediction server - ", e);
    			throw new APIException(APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,e.toString());
    		}
    		catch (Exception e)
            {
    			logger.error("Couldn't retrieve prediction from external prediction server - ", e);
    			throw new APIException(APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,e.toString());
            }
    		finally
    		{
    			
    		}

    }
	}
}
