package io.seldon.engine.grpc;

import java.util.concurrent.ExecutionException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.SeldonGrpc;

/**
 * Passes gRPC requests on to the engine.
 * @author clive
 *
 */
public class SeldonService extends SeldonGrpc.SeldonImplBase {
    
    protected static Logger logger = LoggerFactory.getLogger(SeldonService.class.getName());
    
    private SeldonGrpcServer server;
    
    public SeldonService(SeldonGrpcServer server) {
        super();
        this.server = server;
    }

    @Override
    public void predict(io.seldon.protos.PredictionProtos.SeldonMessage request,
                io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received predict request");
        try
        {
            responseObserver.onNext(server.getPredictionService().predict(request));
        } catch (InterruptedException e) {
            responseObserver.onError(e);
        } catch (ExecutionException e) {
            responseObserver.onError(e);
        }
        finally {}
        responseObserver.onCompleted();
     }
    
    @Override
    public void sendFeedback(io.seldon.protos.PredictionProtos.Feedback request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received feedback request");
        try
        {
            server.getPredictionService().sendFeedback(request);
            responseObserver.onNext(SeldonMessage.newBuilder().build());
        } catch (InterruptedException e) {
            responseObserver.onError(e);
        } catch (ExecutionException e) {
            responseObserver.onError(e);
        }
        finally {}
        responseObserver.onCompleted();
    }
    
}
