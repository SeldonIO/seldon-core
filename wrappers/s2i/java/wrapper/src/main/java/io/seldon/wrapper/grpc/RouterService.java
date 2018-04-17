package io.seldon.wrapper.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.protos.RouterGrpc;

public class RouterService extends RouterGrpc.RouterImplBase  {
    
    protected static Logger logger = LoggerFactory.getLogger(RouterService.class.getName());
    
    private SeldonGrpcServer server;
    
    public RouterService(SeldonGrpcServer server) {
        super();
        this.server = server;
    }
    
    @Override
    public void route(io.seldon.protos.PredictionProtos.SeldonMessage request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received route request");
        responseObserver.onNext(server.getPredictionService().route(request));
        responseObserver.onCompleted();
    }
    
    @Override
    public void sendFeedback(io.seldon.protos.PredictionProtos.Feedback request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received sendFeedback request");
        responseObserver.onNext(server.getPredictionService().sendFeedback(request));
        responseObserver.onCompleted();
    }
    
   
    
}

