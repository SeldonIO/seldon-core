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
package io.seldon.apife.service;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.concurrent.TimeUnit;

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

import io.seldon.apife.AppProperties;
import io.seldon.apife.config.AnnotationsConfig;
import io.seldon.apife.exception.SeldonAPIException;
import io.seldon.protos.DeploymentProtos.Endpoint;

@Service
public class InternalPredictionService {
	
	private static Logger logger = LoggerFactory.getLogger(InternalPredictionService.class.getName());

	private PoolingHttpClientConnectionManager connectionManager;
    private CloseableHttpClient httpClient;
    private final AppProperties appProperties;
 
    public final static String ANNOTATION_REST_CONNECTION_TIMEOUT = "seldon.io/rest-connection-timeout";
    public final static String ANNOTATION_REST_READ_TIMEOUT = "seldon.io/rest-read-timeout";
    public final static String ANNOTATION_GRPC_READ_TIMEOUT = "seldon.io/grpc-read-timeout";

	private static final int DEFAULT_CONNECTION_TIMEOUT = 5000;
	private static final int DEFAULT_READ_TIMEOUT = 10000;
    


    @Autowired
    public InternalPredictionService(AppProperties appProperties, AnnotationsConfig annotations){
    	int connectionTimeout = DEFAULT_CONNECTION_TIMEOUT;
    	if (annotations.has(ANNOTATION_REST_CONNECTION_TIMEOUT))
    	{
    		try
    		{
    			logger.info("Setting REST connection timeout from annotation {}",ANNOTATION_REST_CONNECTION_TIMEOUT);
    			connectionTimeout = Integer.parseInt(annotations.get(ANNOTATION_REST_CONNECTION_TIMEOUT));
    		}
    		catch(NumberFormatException e)
    		{
    			logger.error("Failed to parse REST connection timeout annotation {} with value {}",ANNOTATION_REST_CONNECTION_TIMEOUT,annotations.get(ANNOTATION_REST_CONNECTION_TIMEOUT));
    		}
    	}
    	logger.info("REST Connection timeout set to {}",connectionTimeout);
    	int readTimeout = DEFAULT_READ_TIMEOUT;
    	if (annotations.has(ANNOTATION_REST_READ_TIMEOUT))
    	{
    		try
    		{
    			logger.info("Setting REST read timeout from annotation {}",ANNOTATION_REST_READ_TIMEOUT);
    			readTimeout = Integer.parseInt(annotations.get(ANNOTATION_REST_READ_TIMEOUT));
    		}
    		catch(NumberFormatException e)
    		{
    			logger.error("Failed to parse REST read timeout annotation {} with value {}",ANNOTATION_REST_READ_TIMEOUT,annotations.get(ANNOTATION_REST_READ_TIMEOUT));
    		}
    	}
    	logger.info("REST read timeout set to {}",readTimeout);
    	
    	
        this.appProperties = appProperties;
        connectionManager = new PoolingHttpClientConnectionManager(10,TimeUnit.SECONDS);
        connectionManager.setMaxTotal(150);
        connectionManager.setDefaultMaxPerRoute(150);
        
        RequestConfig requestConfig = RequestConfig.custom()
                .setConnectionRequestTimeout(connectionTimeout)
                .setConnectTimeout(connectionTimeout)
                .setSocketTimeout(readTimeout).build();
        
        httpClient = HttpClients.custom()
                .setConnectionManager(connectionManager)
                .setDefaultRequestConfig(requestConfig)
                .setRetryHandler(new HttpRetryHandler())
                .build();
    }
		
	public String getPrediction(String request,String serviceName) {
		
		return predictREST(request,serviceName);
				
	}
	
	public void sendFeedback(String feedback,String serviceName) {
		sendFeedbackREST(feedback,serviceName);
	}
	
	public void sendFeedbackREST(String feedback,String serviceName) {
		long timeNow = System.currentTimeMillis();
		URI uri;
		try {
			URIBuilder builder = new URIBuilder().setScheme("http")
					.setHost(serviceName)
					.setPort(appProperties.getEngineContainerPort())
					.setPath("/api/v0.1/feedback");

			uri = builder.build();
		} catch (URISyntaxException e) 
		{
			throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_INVALID_ENDPOINT_URL,"Host: "+serviceName+" port:"+appProperties.getEngineContainerPort());
		}
		
		StringEntity requestEntity = new StringEntity(feedback,ContentType.APPLICATION_JSON);
		
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
				resp.getEntity();
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
			throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_MICROSERVICE_ERROR,e.toString());
		}
		catch (Exception e)
        {
			logger.error("Couldn't retrieve prediction from external prediction server - ", e);
			throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_MICROSERVICE_ERROR,e.toString());
        }
		finally
		{
			
		}
	}
	
	
	public String predictREST(String dataString, String serviceName){
		{
    		long timeNow = System.currentTimeMillis();
    		URI uri;
			try {
    			URIBuilder builder = new URIBuilder().setScheme("http")
    					.setHost(serviceName)
    					.setPort(appProperties.getEngineContainerPort())
    					.setPath("/api/v0.1/predictions");

    			uri = builder.build();
    		} catch (URISyntaxException e) 
    		{
    			throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_INVALID_ENDPOINT_URL,"Host: "+serviceName+" port:"+appProperties.getEngineContainerPort());
    		}
			
			StringEntity requestEntity = new StringEntity(dataString,ContentType.APPLICATION_JSON);
			
    		HttpContext context = HttpClientContext.create();
    		HttpPost httpPost = new HttpPost(uri);
    		httpPost.setEntity(requestEntity);
    		try  
    		{
    			//if (logger.isDebugEnabled())
    			logger.info("Requesting " + httpPost.getURI().toString());
    			CloseableHttpResponse resp = httpClient.execute(httpPost, context);
    			try
    			{
    				if (resp.getStatusLine().getStatusCode() != 200)
    				{
        				logger.warn("Received response with code "+resp.getStatusLine().getStatusCode()+" with reason "+resp.getStatusLine().getReasonPhrase());
        				throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_MICROSERVICE_ERROR,String.format("Status code: %s Reason: %s", resp.getStatusLine().getStatusCode(),resp.getStatusLine().getReasonPhrase()));

    				}
    				else
    				{
    					logger.info("Received response");
    					return EntityUtils.toString(resp.getEntity());    	
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
    			throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_MICROSERVICE_ERROR,e.toString());
    		}
    		finally
    		{
    			
    		}

    }
	}
	
	public void setHttpClient(CloseableHttpClient client)
	{
		this.httpClient = client;
	}
}
