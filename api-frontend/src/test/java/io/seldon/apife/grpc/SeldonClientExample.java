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
package io.seldon.apife.grpc;

import java.io.IOException;
import java.net.URISyntaxException;
import java.util.Random;
import java.util.concurrent.TimeUnit;

import org.apache.http.client.ClientProtocolException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;

import io.grpc.CallOptions;
import io.grpc.Channel;
import io.grpc.ClientCall;
import io.grpc.ClientInterceptor;
import io.grpc.ClientInterceptors;
import io.grpc.ForwardingClientCall.SimpleForwardingClientCall;
import io.grpc.ForwardingClientCallListener.SimpleForwardingClientCallListener;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Metadata;
import io.grpc.MethodDescriptor;
import io.seldon.apife.pb.ProtoBufUtils;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.protos.SeldonGrpc;

public class SeldonClientExample {
	protected static Logger logger = LoggerFactory.getLogger(SeldonClientExample.class.getName());
	
	 private final ManagedChannel channel;
	  private final SeldonGrpc.SeldonBlockingStub blockingStub;
	  private final SeldonGrpc.SeldonStub asyncStub;

	  /** Construct client for accessing RouteGuide server at {@code host:port}. */
	  public SeldonClientExample(String host, int port) {
	    this(ManagedChannelBuilder.forAddress(host, port).usePlaintext(true));
	  }

	  /** Construct client for accessing RouteGuide server using the existing channel. */
	  public SeldonClientExample(ManagedChannelBuilder<?> channelBuilder) {
		  ClientInterceptor interceptor = new HeaderClientInterceptor();
	    channel = channelBuilder.build();
	    Channel interceptChannel = ClientInterceptors.intercept(channel, interceptor);
	    blockingStub = SeldonGrpc.newBlockingStub(interceptChannel);
	    asyncStub = SeldonGrpc.newStub(interceptChannel);
	  }

	  public void shutdown() throws InterruptedException {
		    channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
		}
	  
	  public void predict() throws InvalidProtocolBufferException
	  {
		  SeldonMessage request = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addValues(1.0).addShape(1))).build();
		  
		  SeldonMessage response = blockingStub.predict(request);
		  
		  logger.info(ProtoBufUtils.toJson(response));
	  }
	  
	  /** Issues several different requests and then exits. 
	 * @throws InvalidProtocolBufferException */
	  public static void main(String[] args) throws InterruptedException, InvalidProtocolBufferException {

	    SeldonClientExample client = new SeldonClientExample("localhost", SeldonGrpcServer.SERVER_PORT);
	    try {
	    	
	    	client.predict();
	    
	    } finally {
	      client.shutdown();
	    }
	  }
	  
	  /**
	   * A interceptor to handle client header.
	   */
	  public static class HeaderClientInterceptor implements ClientInterceptor {

		  protected static Logger logger = LoggerFactory.getLogger(HeaderClientInterceptor.class.getName());
	    
	    static final Metadata.Key<String> OAUTH_KEY =
	        Metadata.Key.of(HeaderServerInterceptor.OAUTH_TOKEN_HEADER, Metadata.ASCII_STRING_MARSHALLER);

	    @Override
	    public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(MethodDescriptor<ReqT, RespT> method,
	        CallOptions callOptions, Channel next) {
	      return new SimpleForwardingClientCall<ReqT, RespT>(next.newCall(method, callOptions)) {

	        @Override
	        public void start(Listener<RespT> responseListener, Metadata headers) {
	          /* put custom header */
	          OauthTokenProvider p = new OauthTokenProvider();
	          String token;
            try {
                token = p.getToken();
                headers.put(OAUTH_KEY, token);
                super.start(new SimpleForwardingClientCallListener<RespT>(responseListener) {
                  @Override
                  public void onHeaders(Metadata headers) {
                    /**
                     * if you don't need receive header from server,
                     * you can use {@link io.grpc.stub.MetadataUtils#attachHeaders}
                     * directly to send header
                     */
                    logger.info("header received from server:" + headers);
                    super.onHeaders(headers);
                  }
                }, headers);
                
                
            } catch (ClientProtocolException e) {
                // TODO Auto-generated catch block
                e.printStackTrace();
            } catch (URISyntaxException e) {
                // TODO Auto-generated catch block
                e.printStackTrace();
            } catch (IOException e) {
                // TODO Auto-generated catch block
                e.printStackTrace();
            }
	          
	        }
	      };
	    }
	  }

}
