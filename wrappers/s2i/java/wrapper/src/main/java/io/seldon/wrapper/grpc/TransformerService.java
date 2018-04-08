package io.seldon.wrapper.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.protos.TransformerGrpc;

public class TransformerService extends TransformerGrpc.TransformerImplBase  {
    
    protected static Logger logger = LoggerFactory.getLogger(TransformerService.class.getName());
    
    private SeldonGrpcServer server;
    
    public TransformerService(SeldonGrpcServer server) {
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
}
