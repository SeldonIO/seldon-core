package io.seldon.wrapper.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.protos.GenericGrpc;

public class GenericService extends GenericGrpc.GenericImplBase  {
    
    protected static Logger logger = LoggerFactory.getLogger(GenericService.class.getName());
    
    private SeldonGrpcServer server;
    
    public GenericService(SeldonGrpcServer server) {
        super();
        this.server = server;
    }
    
    @Override
    public void transformInput(io.seldon.protos.PredictionProtos.SeldonMessage request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received transformInput request");
        responseObserver.onNext(server.getPredictionService().transformInput(request));
        responseObserver.onCompleted();
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
    
   
    
    @Override
    public void transformOutput(io.seldon.protos.PredictionProtos.SeldonMessage request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received transformOutput request");
        responseObserver.onNext(server.getPredictionService().transformOutput(request));
        responseObserver.onCompleted();
    }

    @Override
    public void aggregate(io.seldon.protos.PredictionProtos.SeldonMessageList request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received aggregate request");
        responseObserver.onNext(server.getPredictionService().aggregate(request));
        responseObserver.onCompleted();
    }
}
