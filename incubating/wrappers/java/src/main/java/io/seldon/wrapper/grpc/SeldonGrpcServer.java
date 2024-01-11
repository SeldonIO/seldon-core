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

import io.grpc.Server;
import io.grpc.netty.NettyServerBuilder;
import io.seldon.wrapper.api.SeldonPredictionService;
import io.seldon.wrapper.config.AnnotationsConfig;
import java.io.IOException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;

@Component
public class SeldonGrpcServer {
  protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServer.class.getName());

  public static final int SERVER_PORT = 5001;
  private final String ANNOTATION_MAX_MESSAGE_SIZE = "seldon.io/grpc-max-message-size";

  private final int port;
  private final Server server;

  private final SeldonPredictionService predictionService;

  @Autowired
  public SeldonGrpcServer(
      AnnotationsConfig annotations,
      SeldonPredictionService predictionService,
      @Value("${grpc.port}") Integer grpcPort) {
    logger.info("grpc port {}", grpcPort);

    port = grpcPort;

    this.predictionService = predictionService;
    NettyServerBuilder builder =
        NettyServerBuilder.forPort(port)
            .addService(new ModelService(this))
            .addService(new RouterService(this))
            .addService(new TransformerService(this))
            .addService(new OutputTransformerService(this))
            .addService(new CombinerService(this))
            .addService(new GenericService(this));
    if (annotations.has(ANNOTATION_MAX_MESSAGE_SIZE)) {
      try {
        int maxMessageSize = Integer.parseInt(annotations.get(ANNOTATION_MAX_MESSAGE_SIZE));
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

  public SeldonPredictionService getPredictionService() {
    return predictionService;
  }

  @Async
  public void runServer() throws InterruptedException, IOException {
    logger.info("Starting grpc server");
    start();
    blockUntilShutdown();
  }

  /** Start serving requests. */
  public void start() throws IOException {
    server.start();
    logger.info("Server started, listening on " + port);
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                // Use stderr here since the logger may has been reset by its JVM shutdown hook.
                System.err.println("*** shutting down gRPC server since JVM is shutting down");
                SeldonGrpcServer.this.stop();
                System.err.println("*** server shut down");
              }
            });
  }

  /** Stop serving requests and shutdown resources. */
  public void stop() {
    if (server != null) {
      server.shutdown();
    }
  }

  /** Await termination on the main thread since the grpc library uses daemon threads. */
  private void blockUntilShutdown() throws InterruptedException {
    if (server != null) {
      server.awaitTermination();
    }
  }
}
