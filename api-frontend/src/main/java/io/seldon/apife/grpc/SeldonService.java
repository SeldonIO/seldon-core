package io.seldon.apife.grpc;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.ManagedChannel;
import io.seldon.protos.DeploymentProtos.DeploymentSpec;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.protos.SeldonGrpc;

public class SeldonService extends SeldonGrpc.SeldonImplBase {
    
    protected static Logger logger = LoggerFactory.getLogger(SeldonService.class.getName());
    
    private SeldonGrpcServer server;
    
    public SeldonService(SeldonGrpcServer server) {
        super();
        this.server = server;
    }

    public void predict(io.seldon.protos.PredictionProtos.SeldonMessage request,
                io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.info("Received request "+server.principalThreadLocal.get());
        ManagedChannel channel = server.getChannel();
        if (channel != null)
        {
            final SeldonGrpc.SeldonBlockingStub blockingStub = SeldonGrpc.newBlockingStub(channel);
            responseObserver.onNext(blockingStub.predict(request));
        }
        else
        {
            SeldonMessage response = SeldonMessage.newBuilder().setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addValues(2.0).addShape(1))).build();
            responseObserver.onNext(response);
        }
        responseObserver.onCompleted();
     }
    
}
