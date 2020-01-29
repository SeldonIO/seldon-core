package io.seldon.wrapper.api;

import com.google.protobuf.InvalidProtocolBufferException;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.wrapper.exception.APIException;
import io.seldon.wrapper.exception.APIException.ApiExceptionType;
import io.seldon.wrapper.pb.ProtoBufUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnExpression;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@ConditionalOnExpression("${seldon.api.transformer.enabled:false}")
public class TransformerRestController {

  private static Logger logger = LoggerFactory.getLogger(TransformerRestController.class.getName());

  @Autowired SeldonPredictionService predictionService;

  @RequestMapping(
      value = "/transform-input",
      method = {RequestMethod.GET, RequestMethod.POST},
      produces = "application/json; charset=utf-8")
  public ResponseEntity<String> transformInput(@RequestParam("json") String json) {
    SeldonMessage request;
    try {
      SeldonMessage.Builder builder = SeldonMessage.newBuilder();
      ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
      request = builder.build();
    } catch (InvalidProtocolBufferException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, json);
    }

    try {
      SeldonMessage response = predictionService.transformInput(request);
      String res = ProtoBufUtils.toJson(response);
      return new ResponseEntity<String>(res, HttpStatus.OK);
    } catch (InvalidProtocolBufferException e) {
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, "");
    }
  }

  @RequestMapping(
      value = "/transform-output",
      method = {RequestMethod.GET, RequestMethod.POST},
      produces = "application/json; charset=utf-8")
  public ResponseEntity<String> transformOutput(@RequestParam("json") String json) {
    SeldonMessage request;
    try {
      SeldonMessage.Builder builder = SeldonMessage.newBuilder();
      ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
      request = builder.build();
    } catch (InvalidProtocolBufferException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, json);
    }

    try {
      SeldonMessage response = predictionService.transformOutput(request);
      String res = ProtoBufUtils.toJson(response);
      return new ResponseEntity<String>(res, HttpStatus.OK);
    } catch (InvalidProtocolBufferException e) {
      throw new APIException(ApiExceptionType.WRAPPER_INVALID_MESSAGE, "");
    }
  }
}
