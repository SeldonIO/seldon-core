package io.seldon.wrapper.api;

/**
 * This is a model prediction API and container everything related to predictions.
 *
 * NOTE:
 * This is NOT a RestFull API, since there is no resource or state (aka there is no prediction object or state)
 * involved.
 */

import com.google.protobuf.InvalidProtocolBufferException;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.wrapper.exception.APIException;
import io.seldon.wrapper.exception.APIException.ApiExceptionType;
import io.seldon.wrapper.pb.ProtoBufUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnExpression;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@ConditionalOnExpression("${seldon.api.model.enabled:false}")
public class ModelPredictionController {
  private static Logger logger = LoggerFactory.getLogger(ModelPredictionController.class.getName());

  @Autowired SeldonPredictionService predictionService;

  /**
   * Will access a POST or a GET request with either a query parameter or a FORM parameter.
   *
   * Examples:
   * GET -> /predict?json={ ... }
   * curl -s \
   *  localhost:9000/predict?json={"data": {"names": ["a", "b"], "ndarray": [[1.0, 2.0]]}}' \
   *
   * POST FORM -> /predict
   * curl -s -X POST \
   *  -d 'json={"data": {"names": ["a", "b"], "ndarray": [[1.0, 2.0]]}}' \
   *  localhost:9000/predict
   *
   * @param json
   * @return
   * @deprecated
   */
  @Deprecated
  @RequestMapping(
      value = "/predict",
      method = {RequestMethod.GET, RequestMethod.POST},
      produces = MediaType.APPLICATION_JSON_UTF8_VALUE
  )
  public ResponseEntity<String> predictLegacy(@RequestParam("json") String json) {
    return this.predict(json);
  }

  @RequestMapping(
      value = "/predict",
      method = {RequestMethod.POST},
      consumes = {
          MediaType.APPLICATION_JSON_VALUE,
          MediaType.APPLICATION_JSON_UTF8_VALUE
      },
      produces = MediaType.APPLICATION_JSON_VALUE
  )
  public ResponseEntity<String> predict(@RequestBody String jsonStr) {
    SeldonMessage request;
    try {
      SeldonMessage.Builder builder = SeldonMessage.newBuilder();
      ProtoBufUtils.updateMessageBuilderFromJson(builder, jsonStr);
      request = builder.build();
    } catch (InvalidProtocolBufferException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, jsonStr);
    }

    try {
      SeldonMessage response = predictionService.predict(request);
      String res = ProtoBufUtils.toJson(response);
      return new ResponseEntity<String>(res, HttpStatus.OK);
    } catch (InvalidProtocolBufferException e) {
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, "");
    }
  }

  @RequestMapping(
      value = "/send-feedback",
      method = {RequestMethod.GET, RequestMethod.POST},
      produces = "application/json; charset=utf-8")
  public ResponseEntity<String> sendFeedback(@RequestParam("json") String json) {
    Feedback request;
    try {
      Feedback.Builder builder = Feedback.newBuilder();
      ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
      request = builder.build();
    } catch (InvalidProtocolBufferException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, json);
    }

    try {
      SeldonMessage response = predictionService.sendFeedback(request);
      String res = ProtoBufUtils.toJson(response);
      return new ResponseEntity<String>(res, HttpStatus.OK);
    } catch (InvalidProtocolBufferException e) {
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, "");
    }
  }
}
