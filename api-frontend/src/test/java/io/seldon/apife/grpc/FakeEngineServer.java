package io.seldon.apife.grpc;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.protos.SeldonGrpc;

public class FakeEngineServer {
    protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServer.class.getName());
    public static final int PORT = 9001;
    private final Server server;
  
    public FakeEngineServer() 
    {
        server = ServerBuilder
                    .forPort(PORT)
                    .addService(new FakeSeldonEngineService())
                    .build();
    }
    
   

  /** Start serving requests. */
    public void start() throws IOException {
      server.start();
      logger.info("Server started, listening on " + PORT);
      Runtime.getRuntime().addShutdownHook(new Thread() {
        @Override
        public void run() {
          // Use stderr here since the logger may has been reset by its JVM shutdown hook.
          System.err.println("*** shutting down gRPC server since JVM is shutting down");
          FakeEngineServer.this.stop();
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
        
        FakeEngineServer server = new FakeEngineServer();
        server.start();
        server.blockUntilShutdown();
  }
    
    public static class FakeSeldonEngineService extends SeldonGrpc.SeldonImplBase {
        
        protected static Logger logger = LoggerFactory.getLogger(SeldonService.class.getName());
        
        public FakeSeldonEngineService() {
            super();
        }

        public void predict(io.seldon.protos.PredictionProtos.SeldonMessage request,
                    io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {

            SeldonMessage response = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addValues(2.0).addShape(1))).build();
            responseObserver.onNext(response);
            responseObserver.onCompleted();
         }
        
    }

    
}
