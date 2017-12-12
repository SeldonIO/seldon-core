package io.seldon.apife.grpc;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.ForwardingServerCall.SimpleForwardingServerCall;
import io.grpc.Metadata;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.ServerCall;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;
import io.grpc.ServerInterceptors;
import io.seldon.protos.ModelGrpc;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;

public class ModelGrpcServer  {
	protected static Logger logger = LoggerFactory.getLogger(ModelGrpcServer.class.getName());
	
	  private final int port;
	  private final Server server;

	  public ModelGrpcServer(int port)
	  {
		  this(ServerBuilder.forPort(port), port);
	  }
	  
	
	  public ModelGrpcServer(ServerBuilder<?> serverBuilder, int port) {
	    this.port = port;
	    server = serverBuilder
	    		.addService(ServerInterceptors.intercept(new ModelService(), new HeaderServerInterceptor()))
	        .build();
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
		
		 public void predict(io.seldon.protos.PredictionProtos.SeldonMessage request,
			        io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
			 SeldonMessage response = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addValues(2.0).addShape(1))).build();
			 logger.info("Received request");
			 responseObserver.onNext(response);
			 responseObserver.onCompleted();
		 }
		
	}
	
	/**
	 * A interceptor to handle server header.
	 */
	public class HeaderServerInterceptor implements ServerInterceptor {


	  final Metadata.Key<String> CUSTOM_HEADER_KEY =
	      Metadata.Key.of("custom_server_header_key", Metadata.ASCII_STRING_MARSHALLER);


	  @Override
	  public <ReqT, RespT> ServerCall.Listener<ReqT> interceptCall(
	      ServerCall<ReqT, RespT> call,
	      final Metadata requestHeaders,
	      ServerCallHandler<ReqT, RespT> next) {
	    logger.info("header received from client:" + requestHeaders);
	    return next.startCall(new SimpleForwardingServerCall<ReqT, RespT>(call) {
	      @Override
	      public void sendHeaders(Metadata responseHeaders) {
	        responseHeaders.put(CUSTOM_HEADER_KEY, "customRespondValue");
	        super.sendHeaders(responseHeaders);
	      }
	    }, requestHeaders);
	  }
	}
	
}
