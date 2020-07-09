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

import io.grpc.Server;
import io.grpc.netty.NettyServerBuilder;
import io.opentracing.contrib.grpc.TracingServerInterceptor;
import io.seldon.engine.config.AnnotationsConfig;
import io.seldon.engine.service.PredictionService;
import io.seldon.engine.tracing.TracingProvider;
import java.io.IOException;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;

@Component
public class SeldonGrpcServer {
  protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServer.class.getName());

  private static final String ENGINE_SERVER_PORT_KEY = "ENGINE_SERVER_GRPC_PORT";
  public static final int SERVER_PORT = 5000;

  public static final String ANNOTATION_MAX_MESSAGE_SIZE = "seldon.io/grpc-max-message-size";

  private final int port;
  private final Server server;

  private final PredictionService predictionService;

  private int maxMessageSize = io.grpc.internal.GrpcUtil.DEFAULT_MAX_MESSAGE_SIZE;

  @Autowired
  public SeldonGrpcServer(
      PredictionService predictionService,
      AnnotationsConfig annotations,
      TracingProvider tracingProvider) {

    { // setup the server port using the env vars
      String engineServerPortString = System.getenv().get(ENGINE_SERVER_PORT_KEY);
      if (engineServerPortString == null) {
        logger.warn(
            "FAILED to find env var [{}], will use defaults for engine server port {}",
            ENGINE_SERVER_PORT_KEY,
            SERVER_PORT);
        port = SERVER_PORT;
      } else {
        port = Integer.parseInt(engineServerPortString);
        logger.info(
            "FOUND env var [{}], will use engine server port {}", ENGINE_SERVER_PORT_KEY, port);
      }
    }
    this.predictionService = predictionService;
    SeldonService seldonService = new SeldonService(this);
    NettyServerBuilder builder;
    if (tracingProvider.isActive()) {
      TracingServerInterceptor tracingInterceptor =
          TracingServerInterceptor.newBuilder().withTracer(tracingProvider.getTracer()).build();
      builder =
          NettyServerBuilder.forPort(port).addService(tracingInterceptor.intercept(seldonService));
    } else {
      builder = NettyServerBuilder.forPort(port).addService(seldonService);
    }
    if (annotations.has(ANNOTATION_MAX_MESSAGE_SIZE)) {
      try {
        maxMessageSize = Integer.parseInt(annotations.get(ANNOTATION_MAX_MESSAGE_SIZE));
        logger.info("Setting max message to {}", maxMessageSize);
        builder.maxInboundMessageSize(maxMessageSize);
      } catch (NumberFormatException e) {
        logger.warn(
            "Failed to parse {} with value {}",
            ANNOTATION_MAX_MESSAGE_SIZE,
            annotations.get(ANNOTATION_MAX_MESSAGE_SIZE),
            e);
      }
    }
    server = builder.build();
  }

  public PredictionService getPredictionService() {
    return predictionService;
  }

  @Async
  public void runServer() throws InterruptedException, IOException {
    logger.info("Starting grpc server");
    start();
  }

  /** Start serving requests. */
  public void start() throws IOException {
    server.start();
    logger.info("Server started, listening on {}", port);
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                // Use stderr here since the logger may has been reset by its JVM shutdown hook.
                System.err.println("*** shutting down gRPC server since JVM is shutting down");
                try {
                  SeldonGrpcServer.this.stop();
                } catch (InterruptedException e) {
                  e.printStackTrace(System.err);
                }
                System.err.println("*** server shut down");
              }
            });
  }

  /** Stop serving requests and shutdown resources. */
  public void stop() throws InterruptedException {
    if (server != null) {
      server.shutdown().awaitTermination(30, TimeUnit.SECONDS);
    }
  }
}
