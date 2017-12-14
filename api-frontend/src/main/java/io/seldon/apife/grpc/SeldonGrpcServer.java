package io.seldon.apife.grpc;

import java.io.IOException;
import java.util.concurrent.ConcurrentHashMap;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.oauth2.provider.token.TokenStore;
import org.springframework.stereotype.Component;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.ServerInterceptors;
import io.seldon.apife.api.oauth.InMemoryClientDetailsService;
import io.seldon.apife.deployments.DeploymentStore;
import io.seldon.apife.exception.SeldonAPIException;
import io.seldon.protos.DeploymentProtos.DeploymentSpec;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonGrpcServer  {
	protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServer.class.getName());
	
	  private final int port;
	  private final Server server;
	  ThreadLocal<String> principalThreadLocal = new ThreadLocal<String>();  
	  ConcurrentHashMap<String,ManagedChannel> channelStore = new ConcurrentHashMap<>();
	  
	  private final DeploymentStore deploymentStore;
	  
	  private final TokenStore tokenStore;
	  
	  @Autowired
	  public SeldonGrpcServer(DeploymentStore deploymentStore,TokenStore tokenStore)
	  {
	      this(deploymentStore,tokenStore,5000);
	  }
	  
	  public SeldonGrpcServer(DeploymentStore deploymentStore,TokenStore tokenStore,int port)
	  {
		  this(deploymentStore,tokenStore,ServerBuilder.forPort(port), port);
	  }
	  
	
	  public SeldonGrpcServer(DeploymentStore deploymentStore,TokenStore tokenStore,ServerBuilder<?> serverBuilder, int port) 
	  {
	      this.deploymentStore = deploymentStore;
	      this.tokenStore = tokenStore;
	      this.port = port;
	      server = serverBuilder
	    		.addService(ServerInterceptors.intercept(new SeldonService(this), new HeaderServerInterceptor(this)))
	        .build();
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
	  
	  

	  public ManagedChannel getChannel() {
	      final String principal = getPrincipal();
          if (principal == null )
          {
              throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_GRPC_NO_PRINCIPAL_FOUND,"");
          }

	      final DeploymentSpec deploymentSpec = deploymentStore.getDeployment(principal);
	      if (deploymentSpec == null)
	      {
	          throw new SeldonAPIException(SeldonAPIException.ApiExceptionType.APIFE_NO_RUNNING_DEPLOYMENT,"");
	      }
	      
	      ManagedChannel channel = channelStore.get(principal);
	      if (channel == null)
	      {
	          Endpoint endPoint = deploymentSpec.getEndpoint();
              channel = ManagedChannelBuilder.forAddress(endPoint.getServiceHost(), endPoint.getServicePort()).usePlaintext(true).build();
	          channelStore.putIfAbsent(principal,channel);
	      }
	      return channel;
    }

    /** Start serving requests. */
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
	   * Main method.  This comment makes the linter happy.
	   */
	  public static void main(String[] args) throws Exception {
	      DeploymentStore store = new DeploymentStore(null,new InMemoryClientDetailsService());
	      SeldonDeployment dep = SeldonDeployment.newBuilder()
	              .setApiVersion("v1alpha1")
	              .setKind("SeldonDeplyment")
	              .setSpec(DeploymentSpec.newBuilder()
	                  .setOauthKey("key")
	                  .setOauthSecret("secret")
	                  .setEndpoint(Endpoint.newBuilder()
	                          .setServiceHost("0.0.0.0")
	                          .setServicePort(FakeEngineServer.PORT))).build();   
	      store.deploymentAdded(dep);
	      SeldonGrpcServer server = new SeldonGrpcServer(store,null,8980);
	      server.start();
	      server.blockUntilShutdown();
	}
}
