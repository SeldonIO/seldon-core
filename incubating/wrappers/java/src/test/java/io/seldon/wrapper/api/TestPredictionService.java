package io.seldon.wrapper.api;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import org.springframework.stereotype.Component;

@Component
public class TestPredictionService implements SeldonPredictionService {
  @Override
  public SeldonMessage predict(SeldonMessage payload) {
    return SeldonMessage.newBuilder()
        .setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addShape(1).addValues(1.0)))
        .build();
  }
}
