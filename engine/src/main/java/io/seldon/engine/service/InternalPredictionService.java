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
import io.seldon.engine.exception.APIException;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.predictors.PredictiveUnitState;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.MABGrpc;
import io.seldon.protos.MABGrpc.MABBlockingStub;
import io.seldon.protos.PredictionProtos.PredictionFeedbackDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef.RequestOneofCase;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.RouteResponseDef;

@Service
public class InternalPredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(InternalPredictionService.class.getName());

	public static final String UNIT_ID_HEADER = "Seldon-Preditive-Unit-ID"; 
	
    ObjectMapper mapper = new ObjectMapper();
    
    RestTemplate restTemplate;
    
    
    
    @Autowired
    public InternalPredictionService(RestTemplate restTemplate){
    	this.restTemplate = restTemplate;
    	
    }
		
	public PredictionResponseDef getPrediction(PredictionRequestDef request, PredictiveUnitState state) throws JsonProcessingException, IOException{

		final EndpointDef endpoint = state.endpoint;
		switch (endpoint.getType()){
			case REST:
				String dataString = ProtoBufUtils.toJson(request);
				boolean isDefault = false;
				if (request.getRequestOneofCase() == RequestOneofCase.REQUEST)
					isDefault = true;
				return predictREST(dataString, state.name, endpoint, isDefault);
				
			case GRPC:
				
		}
		throw new APIException(APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,"no service available");
	}
	
	public RouteResponseDef getRouting(PredictionRequestDef request, EndpointDef endpoint){
		switch (endpoint.getType()){
			case REST:
				throw new NotImplementedException();
			case GRPC:
				return getRoutingGRPC(request, endpoint);
				
		}
		return null;
	}
	
	public void sendFeedback(PredictionFeedbackDef feedback, EndpointDef endpoint){
		switch (endpoint.getType()){
			case REST:
				throw new NotImplementedException();
			case GRPC:
				sendFeedbackGRPC(feedback, endpoint);
		}
		return;
	}
	
	private void sendFeedbackGRPC(PredictionFeedbackDef feedback, EndpointDef endpoint){
		ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();
		MABBlockingStub stub =  MABGrpc.newBlockingStub(channel).withDeadlineAfter(5, TimeUnit.SECONDS);
		
		stub.train(feedback);
		
		return;
	}
	
	private RouteResponseDef getRoutingGRPC(PredictionRequestDef request, EndpointDef endpoint){
		ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint.getServiceHost(), endpoint.getServicePort()).usePlaintext(true).build();
		MABBlockingStub stub =  MABGrpc.newBlockingStub(channel).withDeadlineAfter(5, TimeUnit.SECONDS);
		
		RouteResponseDef routing = stub.route(request);
		return routing;
	}
	
	public PredictionResponseDef predictREST(String dataString, String unitId, EndpointDef endpoint,boolean isDefault){
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
    			headers.add(UNIT_ID_HEADER, unitId);

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
    				    PredictionResponseDef.Builder builder = PredictionResponseDef.newBuilder();
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
