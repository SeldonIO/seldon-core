package io.seldon.apife.service;

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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import com.fasterxml.jackson.core.JsonFactory;
import com.fasterxml.jackson.core.JsonParser;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import io.seldon.apife.exception.APIException;
import io.seldon.protos.DeploymentProtos.EndpointDef;

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
		
	public JsonNode getPrediction(JsonNode requestJson, EndpointDef endpoint) {
		JsonNode ret = null;
		switch (endpoint.getType()){
			case REST:
				
				ret = predictREST(requestJson.toString(), true, endpoint);
				
				break;
			case GRPC:
				
		}
		return ret;
	}
	
	
	
	public JsonNode predictREST(String dataString, Boolean isDefault, EndpointDef endpoint){
		{
    		long timeNow = System.currentTimeMillis();
    		URI uri;
			try {
    			URIBuilder builder = new URIBuilder().setScheme("http")
    					.setHost(endpoint.getServiceHost())
    					.setPort(endpoint.getServicePort())
    					.setPath("/predict")
    					.setParameter("isDefault", isDefault.toString())
    					.setParameter("json", dataString);

    			uri = builder.build();
    		} catch (URISyntaxException e) 
    		{
    			throw new APIException(APIException.INVALID_ENDPOINT_URL);
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
    					ObjectMapper mapper = new ObjectMapper();
    				    JsonFactory factory = mapper.getFactory();
    				    JsonParser parser = factory.createParser(resp.getEntity().getContent());
    				    JsonNode actualObj = mapper.readTree(parser);
    				    
    				    return actualObj;
    				} 
    				else 
    				{
    					logger.error("Couldn't retrieve prediction from external prediction server -- bad http return code: " + resp.getStatusLine().getStatusCode());
    					throw new APIException(APIException.MICROSERVICE_ERROR);
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
    			throw new APIException(APIException.MICROSERVICE_ERROR);
    		}
    		catch (Exception e)
            {
    			logger.error("Couldn't retrieve prediction from external prediction server - ", e);
    			throw new APIException(APIException.MICROSERVICE_ERROR);
            }
    		finally
    		{
    			
    		}

    }
	}
}
