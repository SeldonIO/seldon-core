package io.seldon.wrapper.api;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.SeldonMessageList;

public interface SeldonPredictionService {
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
