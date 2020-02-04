package io.seldon.wrapper.grpc;

import io.seldon.protos.TransformerGrpc;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TransformerService extends TransformerGrpc.TransformerImplBase {

  protected static Logger logger = LoggerFactory.getLogger(TransformerService.class.getName());

  private SeldonGrpcServer server;

  public TransformerService(SeldonGrpcServer server) {
    super();
    this.server = server;
  }

  @Override
  public void transformInput(
      io.seldon.protos.PredictionProtos.SeldonMessage request,
      io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage>
          responseObserver) {
    logger.debug("Received transformInput request");
    responseObserver.onNext(server.getPredictionService().transformInput(request));
    responseObserver.onCompleted();
  }
}
