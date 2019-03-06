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
import java.util.concurrent.ConcurrentHashMap;

import javax.annotation.PostConstruct;

import org.apache.commons.lang.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.security.oauth2.provider.token.TokenStore;
import org.springframework.stereotype.Component;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.ServerInterceptors;
import io.grpc.netty.NettyServerBuilder;
import io.seldon.apife.AppProperties;
import io.seldon.apife.config.AnnotationsConfig;
import io.seldon.apife.deployments.DeploymentStore;
import io.seldon.apife.deployments.DeploymentsHandler;
import io.seldon.apife.exception.SeldonAPIException;
import io.seldon.apife.k8s.KubernetesUtil;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonGrpcServer    {
    protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServer.class.getName());
	
    public static final int SERVER_PORT = 5000;
    private final String ANNOTATION_MAX_MESSAGE_SIZE = "seldon.io/grpc-max-message-size";
    private final String ANNOTATION_GRPC_READ_TIMEOUT = "seldon.io/grpc-read-timeout";
    
    private final int port;
    private final Server server;
    private ThreadLocal<String> principalThreadLocal = new ThreadLocal<String>();  
    private ConcurrentHashMap<String,ManagedChannel> channelStore = new ConcurrentHashMap<>();
	  
    private final DeploymentStore deploymentStore;
    private final TokenStore tokenStore;
    private final AppProperties appProperties;
    private final grpcDeploymentsListener grpcDeploymentsListener;
    private final DeploymentsHandler deploymentsHandler;

    public static final int DEFAULT_GRPC_READ_TIMEOUT = 5000;    
    public static final int TIMEOUT = 5;
    private int maxMessageSize = io.grpc.internal.GrpcUtil.DEFAULT_MAX_MESSAGE_SIZE;
    private int grpcReadTimeout = DEFAULT_GRPC_READ_TIMEOUT;
    private final KubernetesUtil k8sUtil = new KubernetesUtil();
    
    @Autowired
    public SeldonGrpcServer(AppProperties appProperties,DeploymentStore deploymentStore,TokenStore tokenStore,DeploymentsHandler deploymentsHandler,AnnotationsConfig annotations)
    {
        this(appProperties,deploymentStore,tokenStore,deploymentsHandler,annotations,SERVER_PORT);  
    }    
    
    public SeldonGrpcServer(AppProperties appProperties,DeploymentStore deploymentStore,TokenStore tokenStore,DeploymentsHandler deploymentsHandler,AnnotationsConfig annotations,int port)
    {
        this(appProperties,deploymentStore,tokenStore,ServerBuilder.forPort(port), deploymentsHandler, annotations, port);
    }
    
  
    public SeldonGrpcServer(AppProperties appProperties,DeploymentStore deploymentStore,TokenStore tokenStore,ServerBuilder<?> serverBuilder,DeploymentsHandler deploymentsHandler, AnnotationsConfig annotations, int port) 
    {
        this.appProperties = appProperties;
        this.deploymentStore = deploymentStore;
        this.tokenStore = tokenStore;
        this.grpcDeploymentsListener = new grpcDeploymentsListener(this);
        this.deploymentsHandler = deploymentsHandler;
        deploymentsHandler.addListener(this.grpcDeploymentsListener);
        this.port = port;
        NettyServerBuilder builder = NettyServerBuilder
                .forPort(port)
                .addService(ServerInterceptors.intercept(new SeldonService(this), new HeaderServerInterceptor(this)));
        if (annotations != null && annotations.has(ANNOTATION_MAX_MESSAGE_SIZE))
        {
        	try 
        	{
        		maxMessageSize =Integer.parseInt(annotations.get(ANNOTATION_MAX_MESSAGE_SIZE));
        		logger.info("Setting max message to {}",maxMessageSize);
        		builder.maxMessageSize(maxMessageSize);
        	}
        	catch(NumberFormatException e)
        	{
        		logger.warn("Failed to parse {} with value {}",ANNOTATION_MAX_MESSAGE_SIZE,annotations.get(ANNOTATION_MAX_MESSAGE_SIZE),e);        		
        	}
        }
    	if (annotations.has(ANNOTATION_GRPC_READ_TIMEOUT))
        {
        	try 
        	{
        		grpcReadTimeout = Integer.parseInt(annotations.get(ANNOTATION_GRPC_READ_TIMEOUT));
        		logger.info("Setting grpc read timeout to {}ms",grpcReadTimeout);
        	}
        	catch(NumberFormatException e)
        	{
        		logger.error("Failed to parse {} with value {}",ANNOTATION_GRPC_READ_TIMEOUT,annotations.get(ANNOTATION_GRPC_READ_TIMEOUT),e);
        	}
        }
    	logger.info("gRPC read timeout set to {}",grpcReadTimeout);

        server = builder.build();
    }
  
    @PostConstruct
    private void init() throws Exception{
        logger.info("Initializing...");
        deploymentsHandler.addListener(this.grpcDeploymentsListener);
    }
    
    public void setPrincipal(String principal)
    {
        this.principalThreadLocal.set(principal);
    }
    
    
    private String getPrincipal()
    {
        return this.principalThreadLocal.get();
    }
    
    public TokenStore getTokenStore()
    {
        return tokenStore;
    }

    /**
     * Using the principal from authorization return a client gRPC channel to connect to the engine running the prediction graph.
     * @return ManagedChannel
     */
    public ManagedChannel getChannel() {
        final String principal = getPrincipal();
        if (principal == null )
        {
            throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_GRPC_NO_PRINCIPAL_FOUND,"");
        }

        final SeldonDeployment mlDep = deploymentStore.getDeployment(principal);
        if (mlDep == null)
        {
            throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_NO_RUNNING_DEPLOYMENT,"Principal is "+principal);
        }
        
        ManagedChannel channel = channelStore.get(principal);
        if (channel == null)
        {
            throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_GRPC_NO_GRPC_CHANNEL_FOUND,"Principal is "+principal);
        }
        return channel;
    }

    @Async
    public void runServer() throws InterruptedException, IOException
    {
        logger.info("Starting grpc server");
        start();
        blockUntilShutdown();
    }
    
    /** 
     * Start serving requests. 
     */
    public void start() throws IOException {
      server.start();
      logger.info("Server started, listening on " + port);
      Runtime.getRuntime().addShutdownHook(new Thread() {
        @Override
        public void run() {
          // Use stderr here since the logger may has been reset by its JVM shutdown hook.
          System.err.println("*** shutting down gRPC server since JVM is shutting down");
          SeldonGrpcServer.this.stop();
          System.err.println("*** server shut down");
        }
      });
    }

    /** Stop serving requests and shutdown resources. */
    public void stop() {
      if (server != null) {
        server.shutdown();
      }
    }

    /**
     * Await termination on the main thread since the grpc library uses daemon threads.
     */
    private void blockUntilShutdown() throws InterruptedException {
      if (server != null) {
        server.awaitTermination();
      }
    }

    public void deploymentAdded(SeldonDeployment resource) {
    	if (StringUtils.isEmpty(resource.getSpec().getOauthKey()))
    	{
    		logger.warn("Empty oauth key ignoring for {}",resource.getSpec().getName());
    	}
    	else
    	{
    		final String namespace = k8sUtil.getNamespace(resource);
    		final String endpoint = k8sUtil.getSeldonId(resource) + "." + namespace; 
            final ManagedChannel channel = ManagedChannelBuilder.forAddress(endpoint, appProperties.getEngineGrpcContainerPort()).usePlaintext(true).build();
            if (appProperties.isSingleNamespace())
            	channelStore.put(resource.getSpec().getOauthKey(),channel);
            final String namespacedKey = resource.getSpec().getOauthKey() + namespace;
            channelStore.put(namespacedKey,channel);
    	}
    }

    public void deploymentRemoved(SeldonDeployment resource) {
       channelStore.remove(resource.getSpec().getOauthKey());
    }

	public int getMaxMessageSize() {
		return maxMessageSize;
	}
    
    public int getReadTimeout() {
    	return grpcReadTimeout;
    }
   
}
