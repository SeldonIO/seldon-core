/**
 * ***************************************************************************** Copyright 2017
 * Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * <p>Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
 * except in compliance with the License. You may obtain a copy of the License at
 *
 * <p>http://www.apache.org/licenses/LICENSE-2.0
 *
 * <p>Unless required by applicable law or agreed to in writing, software distributed under the
 * License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 * *****************************************************************************
 */
package io.seldon.engine.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.predictors.EnginePredictor;
import io.seldon.engine.predictors.PredictorBean;
import io.seldon.engine.predictors.PredictorState;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import java.io.IOException;
import java.math.BigInteger;
import java.security.SecureRandom;
import java.time.ZonedDateTime;
import java.util.concurrent.ExecutionException;
import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpEntity;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;

@Service
public class PredictionService {

  private static Logger logger = LoggerFactory.getLogger(PredictionService.class.getName());

  @Autowired PredictorBean predictorBean;

  @Autowired EnginePredictor enginePredictor;

  PuidGenerator puidGenerator = new PuidGenerator();

  @Value("${log.requests}")
  private boolean logRequests;

  @Value("${log.responses}")
  private boolean logResponses;

  @Value("${log.feedback.requests}")
  private boolean logFeedbackRequests;

  @Value("${log.messages.externally}")
  private boolean logMessagesExternally;

  @Value("${message.logging.service}")
  private String messageLoggingService;

  public final class PuidGenerator {
    private SecureRandom random = new SecureRandom();

    public String nextPuidId() {
      return new BigInteger(130, random).toString(32);
    }
  }

  public void sendFeedback(Feedback feedback)
      throws InterruptedException, ExecutionException, InvalidProtocolBufferException {
    PredictorState predictorState =
        predictorBean.predictorStateFromPredictorSpec(enginePredictor.getPredictorSpec());

    predictorBean.sendFeedback(feedback, predictorState);

    if (logFeedbackRequests) {
      logMessageAsJson(feedback);
    }

    return;
  }

  public SeldonMessage predict(SeldonMessage request)
      throws InterruptedException, ExecutionException, InvalidProtocolBufferException {

    ZonedDateTime requestTime = ZonedDateTime.now();

    if (!request.hasMeta()) {
      request =
          request
              .toBuilder()
              .setMeta(Meta.newBuilder().setPuid(puidGenerator.nextPuidId()).build())
              .build();
    } else if (StringUtils.isEmpty(request.getMeta().getPuid())) {
      request =
          request
              .toBuilder()
              .setMeta(request.getMeta().toBuilder().setPuid(puidGenerator.nextPuidId()).build())
              .build();
    }
    String puid = request.getMeta().getPuid();

    PredictorState predictorState =
        predictorBean.predictorStateFromPredictorSpec(enginePredictor.getPredictorSpec());

    SeldonMessage predictorReturn = predictorBean.predict(request, predictorState);

    SeldonMessage.Builder builder =
        SeldonMessage.newBuilder(predictorReturn)
            .setMeta(Meta.newBuilder(predictorReturn.getMeta()).setPuid(puid));

    SeldonMessage response = builder.build();

    // raw logging in engine, if enabled
    if (logRequests) {
      // log json now we've added puid
      logMessageAsJson(request);
    }
    if (logResponses) {
      logMessageAsJson(response);
    }

    // enriched logging outside engine, if enabled
    if (logMessagesExternally) {
      ZonedDateTime responseTime = ZonedDateTime.now();
      sendMessagePairAsJson(request, response, requestTime, responseTime);
    }

    return response;
  }

  private JsonNode combineRequestResponse(
      String request, String response, ZonedDateTime requestTime, ZonedDateTime responseTime)
      throws IOException {
    ObjectMapper mapper = new ObjectMapper();
    JsonNode requestNode = mapper.readTree(request);
    JsonNode responseNode = mapper.readTree(response);
    ObjectNode combined = mapper.createObjectNode();
    combined.set("request", requestNode);
    combined.set("response", responseNode);
    ((ObjectNode) combined.get("request"))
        .set("date", mapper.readTree(mapper.writeValueAsString(requestTime.toString())));
    ((ObjectNode) combined.get("response"))
        .set("date", mapper.readTree(mapper.writeValueAsString(responseTime.toString())));
    String depName = System.getenv().get("DEPLOYMENT_NAME");
    if (depName != null) {
      combined.set("sdepName", mapper.readTree(mapper.writeValueAsString(depName)));
    }

    String depNamespace = System.getenv().get("DEPLOYMENT_NAMESPACE");
    if (depNamespace != null && depNamespace != "") {
      combined.set("namespace", mapper.readTree(mapper.writeValueAsString(depNamespace)));
    }

    return combined;
  }

  private void sendMessagePairAsJson(
      SeldonMessage request,
      SeldonMessage response,
      ZonedDateTime requestTime,
      ZonedDateTime responseTime) {
    try {
      String requestJson = ProtoBufUtils.toJson(request);
      String responseJson = ProtoBufUtils.toJson(response);
      JsonNode pair = combineRequestResponse(requestJson, responseJson, requestTime, responseTime);

      MultiValueMap<String, String> headers = new LinkedMultiValueMap<String, String>();
      headers.add("Content-Type", "application/json");
      headers.add("X-B3-Flags", "1");
      headers.add("CE-SpecVersion", "0.2");
      headers.add("CE-Type", "seldon.message.pair");
      headers.add("CE-Time", requestTime.toString());
      headers.add("CE-EventID", request.getMeta().getPuid());

      String depName = System.getenv().get("DEPLOYMENT_NAME");
      if (depName != null) {
        headers.add("CE-Source", "application/json");
      } else {
        headers.add("CE-Source", "seldon");
      }

      HttpEntity<?> requestBody = new HttpEntity<Object>(pair.toString(), headers);
      RestTemplate restTemplate = new RestTemplate();

      restTemplate.postForEntity(messageLoggingService, requestBody, String.class);

    } catch (Exception ex) {
      logger.error("Unable to parse message", ex);
    }
  }

  private void logMessageAsJson(Feedback message) {
    try {
      String json = ProtoBufUtils.toJson(message);
      System.out.println(json);
    } catch (Exception ex) {
      logger.error("Unable to parse message", ex);
    }
  }

  private void logMessageAsJson(Message message) {
    try {
      String json = ProtoBufUtils.toJson(message);
      System.out.println(json);
    } catch (Exception ex) {
      logger.error("Unable to parse message", ex);
    }
  }
}
