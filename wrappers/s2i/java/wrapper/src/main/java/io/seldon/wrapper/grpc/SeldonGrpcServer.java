/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.wrapper.grpc;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.seldon.wrapper.api.SeldonPredictionService;

@Component
public class SeldonGrpcServer  {
    protected static Logger logger = LoggerFactory.getLogger(SeldonGrpcServer.class.getName());
	
    public static final int SERVER_PORT = 5001;
    
    private final int port;
    private final Server server;
	  
    private final SeldonPredictionService predictionService;
    
    @Autowired
    public SeldonGrpcServer(SeldonPredictionService predictionService,@Value("${grpc.port}") Integer grpcPort)
    {
    	logger.info("grpc port {}",grpcPort);
        
    	port = grpcPort;
          
        this.predictionService = predictionService;
        server = ServerBuilder
                .forPort(port)
                .addService(new ModelService(this))
                .addService(new RouterService(this))
                .addService(new TransformerService(this))
                .addService(new OutputTransformerService(this))
                .addService(new CombinerService(this))
                .addService(new GenericService(this))
          .build();
    }
   
    
    
    public SeldonPredictionService getPredictionService() {
        return predictionService;
    }

    @Async
    public void runServer() throws InterruptedException, IOException
    {
        logger.info("Starting grpc server");
        start();
        blockUntilShutdown();
    }
    
    /** 
     * Start serving requests. 
     */
    public void start() throws IOException {
      server.start();
      logger.info("Server started, listening on " + port);
      Runtime.getRuntime().addShutdownHook(new Thread() {
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

    /**
     * Await termination on the main thread since the grpc library uses daemon threads.
     */
    private void blockUntilShutdown() throws InterruptedException {
      if (server != null) {
        server.awaitTermination();
      }
    }

    
}
