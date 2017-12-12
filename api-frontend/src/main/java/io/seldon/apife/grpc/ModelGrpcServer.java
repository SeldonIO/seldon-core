package io.seldon.apife.grpc;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.ServerInterceptors;
import io.seldon.protos.ModelGrpc;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;

public class ModelGrpcServer  {
	protected static Logger logger = LoggerFactory.getLogger(ModelGrpcServer.class.getName());
	
	  private final int port;
	  private final Server server;
	  ThreadLocal<String> principalThreadLocal = new ThreadLocal<String>();  
	  
	  public ModelGrpcServer(int port)
	  {
		  this(ServerBuilder.forPort(port), port);
	  }
	  
	
	  public ModelGrpcServer(ServerBuilder<?> serverBuilder, int port) {
	    this.port = port;
	    server = serverBuilder
	    		.addService(ServerInterceptors.intercept(new ModelService(this), new HeaderServerInterceptor(this)))
	        .build();
	  }
	  
	  public void setPrincipal(String principal)
	  {
	      this.principalThreadLocal.set(principal);
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
	        ModelGrpcServer.this.stop();
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
	    ModelGrpcServer server = new ModelGrpcServer(8980);
	    server.start();
	    server.blockUntilShutdown();
	}
	  
	  
	private static class ModelService extends ModelGrpc.ModelImplBase {
		
	    private ModelGrpcServer server;
	    
	    public ModelService(ModelGrpcServer server) {
            super();
            this.server = server;
        }

	    public void predict(io.seldon.protos.PredictionProtos.SeldonMessage request,
			        io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
			 SeldonMessage response = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addValues(2.0).addShape(1))).build();
			 logger.info("Received request "+server.principalThreadLocal.get());
			 responseObserver.onNext(response);
			 responseObserver.onCompleted();
		 }
		
	}
	
	
	
}
