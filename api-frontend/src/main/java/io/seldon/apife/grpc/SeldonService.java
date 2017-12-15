package io.seldon.apife.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.ManagedChannel;
import io.seldon.apife.exception.SeldonAPIException;
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
        logger.debug("Received predict request for "+server.principalThreadLocal.get());
        try
        {
            ManagedChannel channel = server.getChannel();
            final SeldonGrpc.SeldonBlockingStub blockingStub = SeldonGrpc.newBlockingStub(channel);
            responseObserver.onNext(blockingStub.predict(request));
        }
        catch (SeldonAPIException e)
        {
            responseObserver.onError(e);
        }
        responseObserver.onCompleted();
     }
    
    @Override
    public void sendFeedback(io.seldon.protos.PredictionProtos.Feedback request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received feedback request for "+server.principalThreadLocal.get());
        try
        {
            ManagedChannel channel = server.getChannel();
            final SeldonGrpc.SeldonBlockingStub blockingStub = SeldonGrpc.newBlockingStub(channel);
            responseObserver.onNext(blockingStub.sendFeedback(request));
        }
        catch (SeldonAPIException e)
        {
            responseObserver.onError(e);
        }
        responseObserver.onCompleted();
    }
    
}
