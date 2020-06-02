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
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.protos.SeldonGrpc;
import java.util.concurrent.TimeUnit;

public class SeldonClientExample {
  private final ManagedChannel channel;
  private final SeldonGrpc.SeldonBlockingStub blockingStub;
  private final SeldonGrpc.SeldonStub asyncStub;

  /** Construct client for accessing RouteGuide server at {@code host:port}. */
  public SeldonClientExample(String host, int port) {
    this(ManagedChannelBuilder.forAddress(host, port).usePlaintext(true));
  }

  /** Construct client for accessing RouteGuide server using the existing channel. */
  public SeldonClientExample(ManagedChannelBuilder<?> channelBuilder) {
    channel = channelBuilder.build();
    blockingStub = SeldonGrpc.newBlockingStub(channel);
    asyncStub = SeldonGrpc.newStub(channel);
  }

  public void shutdown() throws InterruptedException {
    channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
  }

  public void predict() throws InvalidProtocolBufferException {
    SeldonMessage request =
        SeldonMessage.newBuilder()
            .setData(
                DefaultData.newBuilder().setTensor(Tensor.newBuilder().addValues(1.0).addShape(1)))
            .build();

    blockingStub.predict(request);
  }

  /**
   * Issues several different requests and then exits.
   *
   * @throws InvalidProtocolBufferException
   */
  public static void main(String[] args)
      throws InterruptedException, InvalidProtocolBufferException {

    SeldonClientExample client = new SeldonClientExample("localhost", SeldonGrpcServer.SERVER_PORT);
    try {

      client.predict();

    } finally {
      client.shutdown();
    }
  }
}
