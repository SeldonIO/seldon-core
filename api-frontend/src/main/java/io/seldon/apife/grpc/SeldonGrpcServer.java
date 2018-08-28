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
import io.seldon.apife.api.oauth.InMemoryClientDetailsService;
import io.seldon.apife.config.AnnotationsConfig;
import io.seldon.apife.deployments.DeploymentStore;
import io.seldon.apife.deployments.DeploymentsHandler;
import io.seldon.apife.exception.SeldonAPIException;
import io.seldon.apife.k8s.DeploymentWatcher;
import io.seldon.protos.DeploymentProtos.DeploymentSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonGrpcServer    {
    protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServer.class.getName());
	
    public static final int SERVER_PORT = 5000;
    private final String ANNOTATION_MAX_MESSAGE_SIZE = "seldon.io/grpc-max-message-size";
    
    private final int port;
    private final Server server;
    private ThreadLocal<String> principalThreadLocal = new ThreadLocal<String>();  
    private ConcurrentHashMap<String,ManagedChannel> channelStore = new ConcurrentHashMap<>();
	  
    private final DeploymentStore deploymentStore;
    private final TokenStore tokenStore;
    private final AppProperties appProperties;
    private final grpcDeploymentsListener grpcDeploymentsListener;
    private final DeploymentsHandler deploymentsHandler;
    
    private int maxMessageSize = io.grpc.internal.GrpcUtil.DEFAULT_MAX_MESSAGE_SIZE;
    
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

        final DeploymentSpec deploymentSpec = deploymentStore.getDeployment(principal);
        if (deploymentSpec == null)
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

    /**
     * Main method for basic testing.
     */
    public static void main(String[] args) throws Exception {
        DeploymentStore store = new DeploymentStore(null,new InMemoryClientDetailsService());
        SeldonDeployment dep = SeldonDeployment.newBuilder()
                .setApiVersion(DeploymentWatcher.VERSION)
                .setKind("SeldonDeplyment")
                .setSpec(DeploymentSpec.newBuilder()
                    .setName("0.0.0.0")
                    .setOauthKey("key")
                    .setOauthSecret("secret")
                    ).build();   
        AppProperties appProperties = new AppProperties();
        appProperties.setEngineGrpcContainerPort(5000);
        store.deploymentAdded(dep);
        SeldonGrpcServer server = new SeldonGrpcServer(appProperties,store,null,null,null,SERVER_PORT);
        server.start();
        server.blockUntilShutdown();
  }

    public void deploymentAdded(SeldonDeployment resource) {
        ManagedChannel channel = ManagedChannelBuilder.forAddress(resource.getSpec().getName(), appProperties.getEngineGrpcContainerPort()).usePlaintext(true).build();
        channelStore.put(resource.getSpec().getOauthKey(),channel);        
    }

    public void deploymentRemoved(SeldonDeployment resource) {
       channelStore.remove(resource.getSpec().getOauthKey());
    }

	public int getMaxMessageSize() {
		return maxMessageSize;
	}
    
    
   
}
