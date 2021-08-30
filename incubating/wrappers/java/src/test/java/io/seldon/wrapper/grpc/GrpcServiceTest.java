package io.seldon.wrapper.grpc;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.wrapper.api.SeldonPredictionService;
import io.seldon.wrapper.api.TestPredictionService;
import io.seldon.wrapper.config.AnnotationsConfig;
import java.io.IOException;
import org.junit.Assert;
import org.junit.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.Mockito;

public class GrpcServiceTest {

  @Test
  public void predictTest() throws IOException {
    AnnotationsConfig aconfig = new AnnotationsConfig();
    SeldonPredictionService predictSvc = new TestPredictionService();
    SeldonGrpcServer grpcServer = new SeldonGrpcServer(aconfig, predictSvc, 9000);
    ModelService modelService = new ModelService(grpcServer);
    io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver =
        mock(io.grpc.stub.StreamObserver.class);

    SeldonMessage msg = SeldonMessage.newBuilder().build();
    modelService.predict(msg, responseObserver);

    ArgumentCaptor<SeldonMessage> argument = ArgumentCaptor.forClass(SeldonMessage.class);
    verify(responseObserver).onNext(argument.capture());
    SeldonMessage response = argument.getAllValues().get(0);
    Assert.assertEquals(0, response.getData().getTensor().getValuesCount());
  }

  @Test
  public void feedbackTest() throws IOException {
    AnnotationsConfig aconfig = new AnnotationsConfig();
    SeldonPredictionService predictSvc = new TestPredictionService();
    SeldonGrpcServer grpcServer = new SeldonGrpcServer(aconfig, predictSvc, 9000);
    ModelService modelService = new ModelService(grpcServer);
    io.grpc.stub.StreamObserver<io.seldon.protos.PredictionProtos.SeldonMessage> responseObserver =
        mock(io.grpc.stub.StreamObserver.class);

    SeldonMessage msg = SeldonMessage.newBuilder().build();
    modelService.predict(msg, responseObserver);

    ArgumentCaptor<SeldonMessage> argument = ArgumentCaptor.forClass(SeldonMessage.class);
    verify(responseObserver, Mockito.times(1)).onNext(argument.capture());
  }
}
