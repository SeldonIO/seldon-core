package io.seldon.apife.service;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;

import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.protocol.HttpClientContext;
import org.apache.http.client.utils.URIBuilder;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;
import org.apache.http.protocol.HttpContext;
import org.apache.http.util.EntityUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

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
		
	public String getPrediction(String request, EndpointDef endpoint) {
		
		return predictREST(request, endpoint);
				
	}
	
	
	
	public String predictREST(String dataString, EndpointDef endpoint){
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
    			throw new APIException(APIException.ApiExceptionType.APIFE_INVALID_ENDPOINT_URL);
    		}
			
			StringEntity requestEntity = new StringEntity(dataString,ContentType.APPLICATION_JSON);
			
    		HttpContext context = HttpClientContext.create();
    		HttpPost httpPost = new HttpPost(uri);
    		httpPost.setEntity(requestEntity);
    		try  
    		{
    			if (logger.isDebugEnabled())
    				logger.debug("Requesting " + httpPost.getURI().toString());
    			CloseableHttpResponse resp = httpClient.execute(httpPost, context);
    			try
    			{
    				if(resp.getStatusLine().getStatusCode() == 200) 
    				{
    					return EntityUtils.toString(resp.getEntity());
    				} 
    				else 
    				{
    					logger.error("Couldn't retrieve prediction from external prediction server -- bad http return code: " + resp.getStatusLine().getStatusCode());
    					throw new APIException(APIException.ApiExceptionType.APIFE_MICROSERVICE_ERROR);
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
    			throw new APIException(APIException.ApiExceptionType.APIFE_MICROSERVICE_ERROR);
    		}
    		catch (Exception e)
            {
    			logger.error("Couldn't retrieve prediction from external prediction server - ", e);
    			throw new APIException(APIException.ApiExceptionType.APIFE_MICROSERVICE_ERROR);
            }
    		finally
    		{
    			
    		}

    }
	}
}
