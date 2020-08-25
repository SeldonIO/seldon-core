package io.seldon.wrapper.api;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.SeldonMessageList;
import io.seldon.wrapper.pb.ProtoBufUtils;
import java.nio.charset.StandardCharsets;

public interface SeldonPredictionService {
  public default byte[] predictRawREST(byte[] request) {
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    String payload = new String(request, StandardCharsets.UTF_8);

    ProtoBufUtils.updateMessageBuilderFromJson(builder, payload);
    SeldonMessage input = builder.build();

    SeldonMessage prediction = this.predict(input);

    String rawPrediction = ProtoBufUtils.toJson(prediction, true);
    return rawPrediction;
  }

  public default byte[] predictRawGRPC(byte[] request) {
    SeldonMessage input = SeldonMessage.parseFrom(request);

    SeldonMessage prediction = this.predict(input);

    byte[] rawPrediction = prediction.toByteArray();
    return rawPrediction;
  }

  public default SeldonMessage predict(SeldonMessage request) {
    return SeldonMessage.newBuilder().build();
  }

  public default SeldonMessage route(SeldonMessage request) {
    return SeldonMessage.newBuilder().build();
  }

  public default SeldonMessage sendFeedback(Feedback request) {
    return SeldonMessage.newBuilder().build();
  }

  public default SeldonMessage transformInput(SeldonMessage request) {
    return SeldonMessage.newBuilder().build();
  }

  public default SeldonMessage transformOutput(SeldonMessage request) {
    return SeldonMessage.newBuilder().build();
  }

  public default SeldonMessage aggregate(SeldonMessageList request) {
    return SeldonMessage.newBuilder().build();
  }
}
