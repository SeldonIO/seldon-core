/**
 * *****************************************************************************
 * 
 * Copyright (c) 2024 Seldon Technologies Ltd.
 * 
 * Use of this software is governed BY
 * (1) the license included in the LICENSE file or
 * (2) if the license included in the LICENSE file is the Business Source License 1.1,
 * the Change License after the Change Date as each is defined in accordance with the LICENSE file.
 *
 * *****************************************************************************
 */
package io.seldon.wrapper.grpc;

import io.seldon.protos.ModelGrpc;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Passes gRPC requests on to the client prediction service.
 *
 * @author clive
 */
public class ModelService extends ModelGrpc.ModelImplBase {

  protected static Logger logger = LoggerFactory.getLogger(ModelService.class.getName());

  private SeldonGrpcServer server;

  public ModelService(SeldonGrpcServer server) {
    super();
    this.server = server;
  }

  @Override
  public void predict(
      io.seldon.protos.PredictionProtos.SeldonMessage request,
      io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage>
          responseObserver) {
    logger.debug("Received predict request");

    responseObserver.onNext(server.getPredictionService().predict(request));
    responseObserver.onCompleted();
  }

  @Override
  public void sendFeedback(
      io.seldon.protos.PredictionProtos.Feedback request,
      io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage>
          responseObserver) {
    logger.debug("Received sendFeedback request");
    responseObserver.onNext(server.getPredictionService().sendFeedback(request));
    responseObserver.onCompleted();
  }
}
