package io.seldon.wrapper.api;

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import io.seldon.wrapper.pb.ProtoBufUtils;
import org.springframework.stereotype.Component;

@Component
public class TestPredictionService implements SeldonPredictionService {
  @Override
  public SeldonMessage predict(SeldonMessage payload) {
    // echo payload back
    return payload.toBuilder().build();
  }
}
