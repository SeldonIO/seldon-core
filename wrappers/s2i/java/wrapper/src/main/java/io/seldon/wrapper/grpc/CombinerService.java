package io.seldon.wrapper.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.protos.CombinerGrpc;

public class CombinerService extends CombinerGrpc.CombinerImplBase  {
    
    protected static Logger logger = LoggerFactory.getLogger(CombinerService.class.getName());
    
    private SeldonGrpcServer server;
    
    public CombinerService(SeldonGrpcServer server) {
        super();
        this.server = server;
    }
    
    @Override
    public void aggregate(io.seldon.protos.PredictionProtos.SeldonMessageList request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received aggregate request");
        responseObserver.onNext(server.getPredictionService().aggregate(request));
        responseObserver.onCompleted();
    }
}