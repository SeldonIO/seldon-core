package io.seldon.engine.predictors;

import com.google.protobuf.InvalidProtocolBufferException;
import io.opentracing.Span;
import io.seldon.protos.PredictionProtos.Metric;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;
import javax.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;

@Component
public class PredictiveUnitBeanProxy {

  private final PredictiveUnitBean predictiveUnitBean;

  @Autowired
  PredictiveUnitBeanProxy(PredictiveUnitBean predictiveUnitBean) {
    this.predictiveUnitBean = predictiveUnitBean;
  }

  @PostConstruct
  public void init() {
    predictiveUnitBean.setProxy(this);
  }

  @Async
  public Future<SeldonMessage> getOutputAsync(
      SeldonMessage input,
      PredictiveUnitState state,
      Map<String, Integer> routingDict,
      Map<String, String> requestPathDict,
      Map<String, List<Metric>> metrics,
      Span activeSpan)
      throws InterruptedException, ExecutionException, InvalidProtocolBufferException {
    return predictiveUnitBean.getOutputAsync(
        input, state, routingDict, requestPathDict, metrics, activeSpan);
  }
}
