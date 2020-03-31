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
package io.seldon.engine.api.rest;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import io.micrometer.core.annotation.Timed;
import io.opentracing.Span;
import io.opentracing.Tracer;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.exception.APIException.ApiExceptionType;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.service.PredictionService;
import io.seldon.engine.tracing.TracingProvider;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import java.io.IOException;
import java.io.InputStream;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.atomic.AtomicBoolean;
import javax.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.RequestEntity;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;
import org.springframework.web.multipart.MultipartHttpServletRequest;

@RestController
public class RestClientController {

  private static Logger logger = LoggerFactory.getLogger(RestClientController.class.getName());

  @Autowired private PredictionService predictionService;

  @Autowired SeldonGraphReadyChecker readyChecker;

  @Autowired TracingProvider tracingProvider;

  private AtomicBoolean ready = new AtomicBoolean(false);

  @PostConstruct
  public void init() {
    ready.set(true);
  }

  @RequestMapping("/")
  String home() {
    return "Hello World!!";
  }

  @RequestMapping(value = "/ping", method = RequestMethod.GET)
  String ping() {
    return "pong";
  }

  @RequestMapping("/ready")
  ResponseEntity<String> ready() {

    HttpHeaders responseHeaders = new HttpHeaders();
    HttpStatus httpStatus;
    String ret;
    if (ready.get() && readyChecker.getReady()) {
      httpStatus = HttpStatus.OK;
      ret = "ready";
    } else {
      logger.warn(
          "Not ready graph checker {}, controller {}", readyChecker.getReady(), ready.get());
      httpStatus = HttpStatus.SERVICE_UNAVAILABLE;
      ret = "Service unavailable";
    }
    ResponseEntity<String> responseEntity =
        new ResponseEntity<String>(ret, responseHeaders, httpStatus);
    return responseEntity;
  }

  @RequestMapping("/live")
  ResponseEntity<String> live() {

    HttpHeaders responseHeaders = new HttpHeaders();
    HttpStatus httpStatus;
    String ret = "live";
    httpStatus = HttpStatus.OK;

    ResponseEntity<String> responseEntity =
        new ResponseEntity<String>(ret, responseHeaders, httpStatus);
    return responseEntity;
  }

  @RequestMapping("/pause")
  String pause() {
    ready.set(false);
    logger.warn("App Paused");
    return "paused";
  }

  @RequestMapping("/unpause")
  String unpause() {
    ready.set(true);
    logger.warn("App UnPaused");
    return "unpaused";
  }
  
  @Timed
  @CrossOrigin(origins = "*")
  @RequestMapping(
          value = {"/api/v1.0/predictions", "/api/v0.1/predictions"},
          method = RequestMethod.POST,
          consumes = "text/*; charset=utf-8",
          produces = "application/json; charset=utf-8")
  public ResponseEntity<String> predictions_text(RequestEntity<String> requestEntity) {
    logger.debug("Received text predict request");
    Span tracingSpan = null;
    if (tracingProvider.isActive()) {
      Tracer tracer = tracingProvider.getTracer();
      tracingSpan = tracer.buildSpan("/api/v0.1/predictions").start();
      tracer.scopeManager().activate(tracingSpan);
    }
    try {
      ObjectMapper mapper = new ObjectMapper();
      Map<String, Object> protoBody = new HashMap<String, Object>() {{
        put("strData", requestEntity.getBody());
      }};
      return _predictions(mapper.writeValueAsString(protoBody));
    } catch (IOException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.REQUEST_IO_EXCEPTION, e.getMessage());
    } finally {
      if (tracingSpan != null) {
        tracingSpan.finish();
      }
    }
  }

  @Timed
  @CrossOrigin(origins = "*")
  @RequestMapping(
          value = {"/api/v1.0/predictions", "/api/v0.1/predictions"},
          method = RequestMethod.POST,
          consumes = "application/octet-stream; charset=utf-8",
          produces = "application/json; charset=utf-8")
  public ResponseEntity<String> predictions_binary(RequestEntity<InputStream> requestEntity) {
    logger.debug("Received binary predict request");
    Span tracingSpan = null;
    if (tracingProvider.isActive()) {
      Tracer tracer = tracingProvider.getTracer();
      tracingSpan = tracer.buildSpan("/api/v0.1/predictions").start();
      tracer.scopeManager().activate(tracingSpan);
    }
    try {
      ObjectMapper mapper = new ObjectMapper();
      Map<String, Object> protoBody = new HashMap<String, Object>() {{
        put("binData", ByteString.readFrom(requestEntity.getBody()));
      }};
      return _predictions(mapper.writeValueAsString(protoBody));
    } catch (IOException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.REQUEST_IO_EXCEPTION, e.getMessage());
    } finally {
      if (tracingSpan != null) {
        tracingSpan.finish();
      }
    }
  }

  @Timed
  @CrossOrigin(origins = "*")
  @RequestMapping(
      value = {"/api/v1.0/predictions", "/api/v0.1/predictions"},
      method = RequestMethod.POST,
      consumes = "application/json; charset=utf-8",
      produces = "application/json; charset=utf-8")
  public ResponseEntity<String> predictions_json(RequestEntity<String> requestEntity) {
    logger.debug("Received predict request");
    Span tracingSpan = null;
    if (tracingProvider.isActive()) {
      Tracer tracer = tracingProvider.getTracer();
      tracingSpan = tracer.buildSpan("/api/v0.1/predictions").start();
      tracer.scopeManager().activate(tracingSpan);
    }
    try {
      return _predictions(requestEntity.getBody());
    } finally {
      if (tracingSpan != null) {
        tracingSpan.finish();
      }
    }
  }

  @Timed
  @CrossOrigin(origins = "*")
  @RequestMapping(
      value = {"/api/v1.0/predictions", "/api/v0.1/predictions"},
      method = RequestMethod.POST,
      consumes = "multipart/form-data",
      produces = "application/json; charset=utf-8")
  public ResponseEntity<String> predictions_multiform(MultipartHttpServletRequest requestEntity) {
    logger.debug("Received predict request");
    Span tracingSpan = null;
    if (tracingProvider.isActive()) {
      Tracer tracer = tracingProvider.getTracer();
      tracingSpan = tracer.buildSpan("/api/v0.1/predictions").start();
      tracer.scopeManager().activate(tracingSpan);
    }
    try {
      ObjectMapper mapper = new ObjectMapper();
      Map<String, Object> mergedParamMap = new HashMap<String, Object>();
      if (requestEntity.getParameterMap() != null) {
        for (Map.Entry<String, String[]> formEntry : requestEntity.getParameterMap().entrySet()) {
          if (formEntry.getKey().equalsIgnoreCase(SeldonMessage.DataOneofCase.STRDATA.name())) {
            mergedParamMap.put(formEntry.getKey(), formEntry.getValue()[0]);
          } else {
            mergedParamMap.put(formEntry.getKey(), mapper.readTree(formEntry.getValue()[0]));
          }
        }
      }
      if (requestEntity.getFileMap() != null) {
        for (Map.Entry<String, MultipartFile> fileEntry : requestEntity.getFileMap().entrySet()) {
          if (fileEntry.getKey().equalsIgnoreCase(SeldonMessage.DataOneofCase.STRDATA.name())) {
            mergedParamMap.put(fileEntry.getKey(), new String(fileEntry.getValue().getBytes()));
          } else {
            mergedParamMap.put(fileEntry.getKey(), fileEntry.getValue().getBytes());
          }
        }
      }

      return _predictions(mapper.writeValueAsString(mergedParamMap));
    } catch (IOException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.REQUEST_IO_EXCEPTION, e.getMessage());

    } finally {
      if (tracingSpan != null) {
        tracingSpan.finish();
      }
    }
  }

  /**
   * It calls the prediction service for the input json. It is the base function for all forms of
   * request Content-type
   *
   * @param json - Input JSON to predict REST api
   * @return The response for prediction service
   */
  private ResponseEntity<String> _predictions(String json) {
    SeldonMessage request;
    try {
      SeldonMessage.Builder builder = SeldonMessage.newBuilder();
      ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
      request = builder.build();
    } catch (InvalidProtocolBufferException e) {
      logger.error("Bad request", e);
      throw new APIException(ApiExceptionType.ENGINE_INVALID_JSON, json);
    }

    try {
      SeldonMessage response = predictionService.predict(request);
      String responseJson = ProtoBufUtils.toJson(response);
      return new ResponseEntity<String>(responseJson, HttpStatus.OK);
    } catch (InterruptedException e) {
      throw new APIException(ApiExceptionType.ENGINE_INTERRUPTED, e.getMessage());
    } catch (ExecutionException e) {
      if (e.getCause().getClass() == APIException.class) {
        throw (APIException) e.getCause();
      } else {
        throw new APIException(ApiExceptionType.ENGINE_EXECUTION_FAILURE, e.getMessage());
      }
    } catch (InvalidProtocolBufferException e) {
      throw new APIException(ApiExceptionType.ENGINE_INVALID_RESPONSE_JSON, "");
    }
  }

  @Timed
  @CrossOrigin(origins = "*")
  @RequestMapping(
      value = {"/api/v1.0/feedback", "/api/v0.1/feedback"},
      method = RequestMethod.POST,
      consumes = "application/json; charset=utf-8",
      produces = "application/json; charset=utf-8")
  public ResponseEntity<String> feedback(RequestEntity<String> requestEntity) {
    logger.debug("Received feedback request");
    Span tracingSpan = null;
    if (tracingProvider.isActive()) {
      Tracer tracer = tracingProvider.getTracer();
      tracingSpan = tracer.buildSpan("/api/v0.1/feedback").start();
      tracer.scopeManager().activate(tracingSpan);
    }
    try {
      Feedback feedback;
      try {
        Feedback.Builder builder = Feedback.newBuilder();
        ProtoBufUtils.updateMessageBuilderFromJson(builder, requestEntity.getBody());
        feedback = builder.build();
      } catch (InvalidProtocolBufferException e) {
        logger.error("Bad request", e);
        throw new APIException(ApiExceptionType.ENGINE_INVALID_JSON, requestEntity.getBody());
      }

      try {
        predictionService.sendFeedback(feedback);
        String json = "{}";
        return new ResponseEntity<String>(json, HttpStatus.OK);
      } catch (InterruptedException e) {
        throw new APIException(ApiExceptionType.ENGINE_INTERRUPTED, e.getMessage());
      } catch (ExecutionException e) {
        if (e.getCause().getClass() == APIException.class) {
          throw (APIException) e.getCause();
        } else {
          throw new APIException(ApiExceptionType.ENGINE_EXECUTION_FAILURE, e.getMessage());
        }
      } catch (InvalidProtocolBufferException e) {
        throw new APIException(ApiExceptionType.ENGINE_INVALID_RESPONSE_JSON, "");
      }
    } finally {
      if (tracingSpan != null) {
        tracingSpan.finish();
      }
    }
  }
}
