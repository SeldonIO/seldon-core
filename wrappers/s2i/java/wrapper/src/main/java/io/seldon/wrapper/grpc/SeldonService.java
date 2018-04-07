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

import java.util.concurrent.ExecutionException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.SeldonGrpc;

/**
 * Passes gRPC requests on to the engine.
 * @author clive
 *
 */
public class SeldonService extends SeldonGrpc.SeldonImplBase {
    
    protected static Logger logger = LoggerFactory.getLogger(SeldonService.class.getName());
    
    private SeldonGrpcServer server;
    
    public SeldonService(SeldonGrpcServer server) {
        super();
        this.server = server;
    }

    @Override
    public void predict(io.seldon.protos.PredictionProtos.SeldonMessage request,
                io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received predict request");
        
        responseObserver.onNext(server.getPredictionService().predict(request));
        responseObserver.onCompleted();
     }
    
    @Override
    public void sendFeedback(io.seldon.protos.PredictionProtos.Feedback request,
            io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver) {
        logger.debug("Received feedback request");
        responseObserver.onNext(server.getPredictionService().sendFeedback(request));
        responseObserver.onCompleted();
    }
    
}
