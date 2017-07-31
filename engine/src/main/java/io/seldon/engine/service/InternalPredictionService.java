package io.seldon.engine.service;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;

import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.protocol.HttpClientContext;
import org.apache.http.client.utils.URIBuilder;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;
import org.apache.http.protocol.HttpContext;
import org.apache.http.util.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.util.JsonFormat;

import io.seldon.engine.exception.APIException;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef.RequestOneofCase;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;

@Service
public class InternalPredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(InternalPredictionService.class.getName());

	private PoolingHttpClientConnectionManager connectionManager;
    private CloseableHttpClient httpClient;
    
    private static final int DEFAULT_REQ_TIMEOUT = 200;
    private static final int DEFAULT_CON_TIMEOUT = 500;
    private static final int DEFAULT_SOCKET_TIMEOUT = 2000;

    ObjectMapper mapper = new ObjectMapper();
    
    @Autowired
    public InternalPredictionService(){
        connectionManager = new PoolingHttpClientConnectionManager();
        connectionManager.setMaxTotal(150);
        connectionManager.setDefaultMaxPerRoute(150);
        
        RequestConfig requestConfig = RequestConfig.custom()
                .setConnectionRequestTimeout(DEFAULT_REQ_TIMEOUT)
                .setConnectTimeout(DEFAULT_CON_TIMEOUT)
                .setSocketTimeout(DEFAULT_SOCKET_TIMEOUT).build();
        
        httpClient = HttpClients.custom()
                .setConnectionManager(connectionManager)
                .setDefaultRequestConfig(requestConfig)
                .build();
    }
		
	public PredictionResponseDef getPrediction(PredictionRequestDef request, EndpointDef endpoint) throws JsonProcessingException, IOException{

		switch (endpoint.getType()){
			case REST:
				String dataString = ProtoBufUtils.toJson(request);
				boolean isDefault = false;
				if (request.getRequestOneofCase() == RequestOneofCase.REQUEST)
					isDefault = true;
				return predictREST(dataString, endpoint, isDefault);
				
			case GRPC:
				
		}
		throw new APIException(APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,"no service available");
	}
	
	
	public PredictionResponseDef predictREST(String dataString, EndpointDef endpoint,boolean isDefault){
		{
    		long timeNow = System.currentTimeMillis();
    		URI uri;
			try {
    			URIBuilder builder = new URIBuilder().setScheme("http")
    					.setHost(endpoint.getServiceHost())
    					.setPort(endpoint.getServicePort())
    					.setPath("/predict")
    					.setParameter("json", dataString)
    					.setParameter("isDefault", Boolean.toString(isDefault));

    			uri = builder.build();
    		} catch (URISyntaxException e) 
    		{
    			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ENDPOINT_URL,"Host: "+endpoint.getServiceHost()+" port:"+endpoint.getServicePort());
    		}
    		HttpContext context = HttpClientContext.create();
    		HttpGet httpGet = new HttpGet(uri);
    		try  
    		{
    			if (logger.isDebugEnabled())
    				logger.debug("Requesting " + httpGet.getURI().toString());
    			CloseableHttpResponse resp = httpClient.execute(httpGet, context);
    			try
    			{
    				if(resp.getStatusLine().getStatusCode() == 200) 
    				{
    				    PredictionResponseDef.Builder builder = PredictionResponseDef.newBuilder();
    				    String response = EntityUtils.toString(resp.getEntity());
    				    logger.info(response);
    				    JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
    				    return builder.build();
    				} 
    				else 
    				{
    					logger.error("Couldn't retrieve prediction from external prediction server -- bad http return code: " + resp.getStatusLine().getStatusCode());
    					throw new APIException(APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,String.format("Bad return code %d", resp.getStatusLine().getStatusCode()));
    				}
    			}
    			finally
    			{
    				if (resp != null)
    					resp.close();
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
