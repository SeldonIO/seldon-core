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
//    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
//    ProtoBufUtils.updateMessageBuilderFromJson(builder, payload.getData());
//    request = builder.build();
//    SeldonMessage response = SeldonMessage.newBuilder();
//    ProtoBufUtils.updateMessageBuilderFromJson(response, payload);
//    return response;
//    return SeldonMessage.newBuilder()
//        .setData(DefaultData.newBuilder().setTensor(Tensor.newBuilder().addShape(1).addValues(1.0)))
//        .build();
  }
}
