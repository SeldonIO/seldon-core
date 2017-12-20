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
import java.io.InterruptedIOException;
import java.net.SocketException;
import java.rmi.UnknownHostException;

import javax.net.ssl.SSLException;

import org.apache.http.HttpEntityEnclosingRequest;
import org.apache.http.HttpRequest;
import org.apache.http.client.HttpRequestRetryHandler;
import org.apache.http.client.protocol.HttpClientContext;
import org.apache.http.conn.ConnectTimeoutException;
import org.apache.http.protocol.HttpContext;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class HttpRetryHandler implements HttpRequestRetryHandler {

	private static Logger logger = LoggerFactory.getLogger(HttpRetryHandler.class.getName());
	@Override
	public boolean retryRequest(IOException exception, int executionCount, HttpContext context) {
		if (executionCount >= 3) {
			logger.info("Got too many exceptions");
            // Do not retry if over max retry count
            return false;
        }
        if (exception instanceof InterruptedIOException) {
            // Timeout
        	logger.info("Got interrupted exception");
            return true;
        }
        if (exception instanceof SocketException)
        {
        	logger.info("Got socket exception");
        	return true;
        }
        if (exception instanceof org.apache.http.NoHttpResponseException){
            logger.warn("got no http response exception");
            return true;
        }
        if (exception instanceof UnknownHostException) {
            // Unknown host
            return false;
        }
        if (exception instanceof ConnectTimeoutException) {
            // Connection refused
            return true;
        }
        if (exception instanceof SSLException) {
            // SSL handshake exception
            return false;
        }
        
        HttpClientContext clientContext = HttpClientContext.adapt(context);
        HttpRequest request = clientContext.getRequest();
        boolean idempotent = !(request instanceof HttpEntityEnclosingRequest);
        if (idempotent) {
            // Retry if the request is considered idempotent
            return true;
        }
        return false;
	}

}
