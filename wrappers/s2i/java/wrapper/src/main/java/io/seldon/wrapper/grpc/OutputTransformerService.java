package io.seldon.wrapper.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.seldon.protos.OutputTransformerGrpc;

public class OutputTransformerService extends OutputTransformerGrpc.OutputTransformerImplBase  {
    
    protected static Logger logger = LoggerFactory.getLogger(OutputTransformerService.class.getName());
    
    private SeldonGrpcServer server;
    
    public OutputTransformerService(SeldonGrpcServer server) {
        super();
        this.server = server;
    }
    
    @Override
    public void transformOutput(io.seldon.protos.PredictionProtos.SeldonMessage request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received transformOutput request");
        responseObserver.onNext(server.getPredictionService().transformOutput(request));
        responseObserver.onCompleted();
    }
}