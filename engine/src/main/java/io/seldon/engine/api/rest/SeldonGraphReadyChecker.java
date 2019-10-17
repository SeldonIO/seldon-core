package io.seldon.engine.api.rest;

import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.predictors.PredictiveUnitBean;
import io.seldon.engine.predictors.PredictiveUnitImpl;
import io.seldon.engine.predictors.PredictiveUnitState;
import io.seldon.engine.predictors.PredictorBean;
import io.seldon.engine.predictors.PredictorConfigBean;
import io.seldon.engine.predictors.PredictorState;
import io.seldon.protos.DeploymentProtos.Endpoint;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.util.concurrent.atomic.AtomicBoolean;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

@Component
public class SeldonGraphReadyChecker extends PredictiveUnitImpl {

  private static final Logger logger = LoggerFactory.getLogger(SeldonGraphReadyChecker.class);
  @Autowired EnginePredictor enginePredictor;

  @Autowired PredictorBean predictorBean;

  @Autowired PredictiveUnitBean pcb;

  @Autowired public PredictorConfigBean predictorConfig;

  private AtomicBoolean ready = new AtomicBoolean();

  public SeldonGraphReadyChecker() {
    ready.set(false);
  }

  public boolean checkReady(PredictiveUnitState state) {
    PredictiveUnitImpl implementation = predictorConfig.getImplementation(state);
    if (implementation == null) {
      implementation = this;
    }

    if (!implementation.ready(state)) {
      logger.info("{} not ready!", state.name);
      return false;
    } else {
      if (state.children.isEmpty()) return true;
      else {
        for (PredictiveUnitState childState : state.children) {
          if (!checkReady(childState)) return false;
        }
        return true;
      }
    }
  }

  public boolean ready(PredictiveUnitState state) {
    if (predictorConfig.hasMethod(state)) return pingHost(state.endpoint);
    else return false;
  }

  private boolean pingHost(Endpoint endpoint) {
    for (int i = 0; i < 3; i++) {
      Socket socket = null;
      try {
        socket = new Socket();
        socket.connect(
            new InetSocketAddress(endpoint.getServiceHost(), endpoint.getServicePort()), 500);
        socket.close();
        socket = null;
        return true;
      } catch (IOException e) {
        logger.warn(
            "Failed to connect to {}:{}", endpoint.getServiceHost(), endpoint.getServicePort());
      } finally {
        if (socket != null) {
          try {
            socket.close();
          } catch (IOException e) {
            logger.error("Failed to close socket", e);
          }
        }
      }
    }
    logger.warn("Failing {}:{}", endpoint.getServiceHost(), endpoint.getServicePort());
    return false;
  }

  public boolean getReady() {
    return ready.get();
  }

  @Scheduled(fixedDelay = 5000)
  public void checkReady() {
    PredictorState predictorState =
        predictorBean.predictorStateFromPredictorSpec(enginePredictor.getPredictorSpec());
    PredictiveUnitState state = predictorState.rootState;
    boolean graphReady = checkReady(state);
    logger.debug("Seldon graph ready: {}", ready);
    ready.set(graphReady);
  }
}
