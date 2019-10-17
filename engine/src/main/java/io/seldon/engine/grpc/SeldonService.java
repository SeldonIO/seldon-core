/**
 * ***************************************************************************** Copyright 2017
 * Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * <p>Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 *
 * <p>http://www.apache.org/licenses/LICENSE-2.0
 *
 * <p>Unless required by applicable law or agreed to in writing, software distributed under the
 * License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 * *****************************************************************************
 */
package io.seldon.engine.grpc;

import com.google.protobuf.InvalidProtocolBufferException;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.SeldonGrpc;
import java.util.concurrent.ExecutionException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Passes gRPC requests on to the engine.
 *
 * @author clive
 */
public class SeldonService extends SeldonGrpc.SeldonImplBase {

  protected static Logger logger = LoggerFactory.getLogger(SeldonService.class.getName());

  private SeldonGrpcServer server;

  public SeldonService(SeldonGrpcServer server) {
    super();
    this.server = server;
  }

  @Override
  public void predict(
      io.seldon.protos.PredictionProtos.SeldonMessage request,
      io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage>
          responseObserver) {
    logger.debug("Received predict request");
    try {
      responseObserver.onNext(server.getPredictionService().predict(request));
    } catch (InterruptedException e) {
      responseObserver.onError(e);
    } catch (ExecutionException e) {
      responseObserver.onError(e);
    } catch (InvalidProtocolBufferException e) {
      responseObserver.onError(e);
    } finally {
    }
    responseObserver.onCompleted();
  }

  @Override
  public void sendFeedback(
      io.seldon.protos.PredictionProtos.Feedback request,
      io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage>
          responseObserver) {
    logger.debug("Received feedback request");
    try {
      server.getPredictionService().sendFeedback(request);
      responseObserver.onNext(SeldonMessage.newBuilder().build());
    } catch (InterruptedException e) {
      responseObserver.onError(e);
    } catch (ExecutionException e) {
      responseObserver.onError(e);
    } catch (InvalidProtocolBufferException e) {
      responseObserver.onError(e);
    } finally {
    }
    responseObserver.onCompleted();
  }
}
