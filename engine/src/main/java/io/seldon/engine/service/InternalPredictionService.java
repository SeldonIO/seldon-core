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

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;
import io.opentracing.contrib.spring.web.client.TracingRestTemplateInterceptor;
import io.seldon.engine.config.AnnotationsConfig;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.grpc.GrpcChannelHandler;
import io.seldon.engine.grpc.SeldonGrpcServer;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.engine.predictors.PredictiveUnitState;
import io.seldon.engine.tracing.TracingProvider;
import io.seldon.protos.CombinerGrpc;
import io.seldon.protos.CombinerGrpc.CombinerBlockingStub;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.GenericGrpc;
import io.seldon.protos.GenericGrpc.GenericBlockingStub;
import io.seldon.protos.ModelGrpc;
import io.seldon.protos.ModelGrpc.ModelBlockingStub;
import io.seldon.protos.OutputTransformerGrpc;
import io.seldon.protos.OutputTransformerGrpc.OutputTransformerBlockingStub;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.SeldonMessage.DataOneofCase;
import io.seldon.protos.PredictionProtos.SeldonMessageList;
import io.seldon.protos.PredictionProtos.Status;
import io.seldon.protos.RouterGrpc;
import io.seldon.protos.RouterGrpc.RouterBlockingStub;
import io.seldon.protos.TransformerGrpc;
import io.seldon.protos.TransformerGrpc.TransformerBlockingStub;
import java.net.URI;
import java.net.URISyntaxException;
import java.time.Duration;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import org.apache.http.client.utils.URIBuilder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.HttpStatusCodeException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;

@Service
public class InternalPredictionService {

  private static Logger logger = LoggerFactory.getLogger(InternalPredictionService.class.getName());

  public static final String MODEL_NAME_HEADER = "Seldon-model-name";
  public static final String MODEL_IMAGE_HEADER = "Seldon-model-image";
  public static final String MODEL_VERSION_HEADER = "Seldon-model-version";

  public static final String ANNOTATION_REST_CONNECTION_TIMEOUT =
      "seldon.io/rest-connection-timeout";
  public static final String ANNOTATION_REST_READ_TIMEOUT = "seldon.io/rest-read-timeout";
  public static final String ANNOTATION_REST_RETRIES = "seldon.io/rest-connect-retries";
  public static final String ANNOTATION_GRPC_READ_TIMEOUT = "seldon.io/grpc-read-timeout";

  private static final int DEFAULT_CONNECTION_TIMEOUT = 200;
  private static final int DEFAULT_READ_TIMEOUT = 5000;

  public static final int DEFAULT_GRPC_READ_TIMEOUT = 5000;
  public static final int DEFAULT_MAX_RETRIES = 3;

  ObjectMapper mapper = new ObjectMapper();

  RestTemplate restTemplate;

  private int grpcMaxMessageSize = io.grpc.internal.GrpcUtil.DEFAULT_MAX_MESSAGE_SIZE;
  private int grpcReadTimeout = DEFAULT_GRPC_READ_TIMEOUT;
  private int restRetries = DEFAULT_MAX_RETRIES;

  private final GrpcChannelHandler grpcChannelHandler;

  private final Map<String, HttpHeaders> headersCache = new ConcurrentHashMap<>();
  private final Map<String, URI> uriCache = new ConcurrentHashMap<>();

  @Autowired
  public InternalPredictionService(
      RestTemplateBuilder restTemplateBuilder,
      AnnotationsConfig annotations,
      GrpcChannelHandler grpcChannelHandler,
      TracingProvider tracingProvider) {
    this.grpcChannelHandler = grpcChannelHandler;
    int connectionTimeout = DEFAULT_CONNECTION_TIMEOUT;
    if (annotations.has(ANNOTATION_REST_CONNECTION_TIMEOUT)) {
      try {
        logger.info(
            "Setting REST connection timeout from annotation {}",
            ANNOTATION_REST_CONNECTION_TIMEOUT);
        connectionTimeout = Integer.parseInt(annotations.get(ANNOTATION_REST_CONNECTION_TIMEOUT));
      } catch (NumberFormatException e) {
        logger.error(
            "Failed to parse REST connection timeout annotation {} with value {}",
            ANNOTATION_REST_CONNECTION_TIMEOUT,
            annotations.get(ANNOTATION_REST_CONNECTION_TIMEOUT));
      }
    }
    logger.info("REST Connection timeout set to {}", connectionTimeout);
    int readTimeout = DEFAULT_READ_TIMEOUT;
    if (annotations.has(ANNOTATION_REST_READ_TIMEOUT)) {
      try {
        logger.info("Setting REST read timeout from annotation {}", ANNOTATION_REST_READ_TIMEOUT);
        readTimeout = Integer.parseInt(annotations.get(ANNOTATION_REST_READ_TIMEOUT));
      } catch (NumberFormatException e) {
        logger.error(
            "Failed to parse REST read timeout annotation {} with value {}",
            ANNOTATION_REST_READ_TIMEOUT,
            annotations.get(ANNOTATION_REST_READ_TIMEOUT));
      }
    }
    logger.info("REST read timeout set to {}", readTimeout);
    this.restTemplate =
        restTemplateBuilder
            .setConnectTimeout(Duration.ofMillis(connectionTimeout))
            .setReadTimeout(Duration.ofMillis(readTimeout))
            .build();
    if (tracingProvider.isActive()) {
      restTemplate.setInterceptors(
          Collections.singletonList(
              new TracingRestTemplateInterceptor(tracingProvider.getTracer())));
    }
    if (annotations.has(SeldonGrpcServer.ANNOTATION_MAX_MESSAGE_SIZE)) {
      try {
        grpcMaxMessageSize =
            Integer.parseInt(annotations.get(SeldonGrpcServer.ANNOTATION_MAX_MESSAGE_SIZE));
        logger.info("Setting max message to {} bytes", grpcMaxMessageSize);
      } catch (NumberFormatException e) {
        logger.error(
            "Failed to parse {} with value {}",
            SeldonGrpcServer.ANNOTATION_MAX_MESSAGE_SIZE,
            annotations.get(SeldonGrpcServer.ANNOTATION_MAX_MESSAGE_SIZE),
            e);
      }
    }
    logger.info("gRPC max message size set to {}", grpcMaxMessageSize);
    if (annotations.has(ANNOTATION_GRPC_READ_TIMEOUT)) {
      try {
        grpcReadTimeout = Integer.parseInt(annotations.get(ANNOTATION_GRPC_READ_TIMEOUT));
        logger.info("Setting grpc read timeout to {}ms", grpcReadTimeout);
      } catch (NumberFormatException e) {
        logger.error(
            "Failed to parse {} with value {}",
            ANNOTATION_GRPC_READ_TIMEOUT,
            annotations.get(ANNOTATION_GRPC_READ_TIMEOUT),
            e);
      }
    }
    logger.info("gRPC read timeout set to {}", grpcReadTimeout);
    if (annotations.has(ANNOTATION_REST_RETRIES)) {
      try {
        restRetries = Integer.parseInt(annotations.get(ANNOTATION_REST_RETRIES));
        logger.info("Setting rest retries to {}", restRetries);
      } catch (NumberFormatException e) {
        logger.error(
            "Failed to parse {} with value {}",
            ANNOTATION_REST_RETRIES,
            annotations.get(ANNOTATION_REST_RETRIES),
            e);
      }
    }
    logger.info("REST retries set to {}", restRetries);
  }

  public SeldonMessage route(SeldonMessage input, PredictiveUnitState state)
      throws InvalidProtocolBufferException {
    final Endpoint endpoint = state.endpoint;
    switch (endpoint.getType()) {
      case REST:
        String dataString = ProtoBufUtils.toJson(input);
        return queryREST("route", dataString, state, endpoint, isDefaultData(input));

      case GRPC:
        if (state.type == PredictiveUnitType.UNKNOWN_TYPE) {
        	try {
        		GenericBlockingStub stub =
        				GenericGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return stub.route(input);
        	} catch (Exception e) {
                logger.error("grpc exception genericStub route", e);
                throw e;
              }
        } else {
        	try {
        		RouterBlockingStub stub =
        				RouterGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return stub.route(input);
        	} catch (Exception e) {
                logger.error("grpc exception routerStub route", e);
                throw e;
              }
        }
    }
    throw new APIException(
        APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "no service available");
  }

  public SeldonMessage sendFeedback(Feedback feedback, PredictiveUnitState state)
      throws InvalidProtocolBufferException {
    final Endpoint endpoint = state.endpoint;
    switch (endpoint.getType()) {
      case REST:
        String dataString = ProtoBufUtils.toJson(feedback);
        return queryREST("send-feedback", dataString, state, endpoint, true);

      case GRPC:
        if (state.type == PredictiveUnitType.UNKNOWN_TYPE) {
        	try {
        		GenericBlockingStub stub =
        				GenericGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return stub.sendFeedback(feedback);
        	} catch (Exception e) {
                logger.error("grpc exception genericStub sendFeedback", e);
                throw e;
              }
        } else if (state.type == PredictiveUnitType.MODEL) {
        	try {
        		ModelBlockingStub modelStub =
        				ModelGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return modelStub.sendFeedback(feedback);
        	} catch (Exception e) {
                logger.error("grpc exception modelStub sendFeedback", e);
                throw e;
              }
        } else {
        	try {
        		RouterBlockingStub routerStub =
        				RouterGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return routerStub.sendFeedback(feedback);
        	} catch (Exception e) {
                logger.error("grpc exception routerStub sendFeedback", e);
                throw e;
              }
        }
    }
    throw new APIException(
        APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "no service available");
  }

  public SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state)
      throws InvalidProtocolBufferException {
    logger.info("Calling grpc for transform-input");
    final Endpoint endpoint = state.endpoint;
    switch (endpoint.getType()) {
      case REST:
        String dataString = ProtoBufUtils.toJson(input);
        if (state.type == PredictiveUnitType.MODEL) {
          return queryREST("predict", dataString, state, endpoint, isDefaultData(input));
        } else {
          return queryREST("transform-input", dataString, state, endpoint, isDefaultData(input));
        }

      case GRPC:
        switch (state.type) {
          case UNKNOWN_TYPE:
        	try {
        	  GenericBlockingStub genStub =
        		GenericGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
                    .withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
                    .withMaxInboundMessageSize(grpcMaxMessageSize)
                    .withMaxOutboundMessageSize(grpcMaxMessageSize);
        	  return genStub.transformInput(input);
        	 } catch (Exception e) {
                 logger.error("grpc exception on genericStub transformInput ", e);
                 throw e;
               }
          case MODEL:
            try {
              ModelBlockingStub modelStub =
                  ModelGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
                      .withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
                      .withMaxInboundMessageSize(grpcMaxMessageSize)
                      .withMaxOutboundMessageSize(grpcMaxMessageSize);
              logger.debug(modelStub.getCallOptions().toString());
              return modelStub.predict(input);
            } catch (Exception e) {
              logger.error("grpc exception on modelStub predict ", e);
              throw e;
            }
          case TRANSFORMER:
            try {
              TransformerBlockingStub transformerStub =
                  TransformerGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
                      .withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
                      .withMaxInboundMessageSize(grpcMaxMessageSize)
                      .withMaxOutboundMessageSize(grpcMaxMessageSize);
              return transformerStub.transformInput(input);
            } catch (Exception e) {
              logger.error("grpc exception transformStub transformInput", e);
              throw e;
            }
          default:
            throw new APIException(
                APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "Unhandled type");
        }
    }
    throw new APIException(
        APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "no service available");
  }

  public SeldonMessage transformOutput(SeldonMessage output, PredictiveUnitState state)
      throws InvalidProtocolBufferException {
    final Endpoint endpoint = state.endpoint;
    switch (endpoint.getType()) {
      case REST:
        String dataString = ProtoBufUtils.toJson(output);
        return queryREST("transform-output", dataString, state, endpoint, isDefaultData(output));

      case GRPC:
        if (state.type == PredictiveUnitType.UNKNOWN_TYPE) {
        	try {
        		GenericBlockingStub stub =
        				GenericGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return stub.transformOutput(output);
        	} catch (Exception e) {
                logger.error("grpc exception genericStub transformOutput", e);
                throw e;
              }
        } else {
        	try {
        		OutputTransformerBlockingStub stub =
        				OutputTransformerGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return stub.transformOutput(output);
        	} catch (Exception e) {
                logger.error("grpc exception outputTransformerStub transformOutput", e);
                throw e;
              }
        }
    }
    throw new APIException(
        APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "no service available");
  }

  public SeldonMessage aggregate(List<SeldonMessage> outputs, PredictiveUnitState state)
      throws InvalidProtocolBufferException {
    final Endpoint endpoint = state.endpoint;
    SeldonMessageList outputsList =
        SeldonMessageList.newBuilder().addAllSeldonMessages(outputs).build();
    switch (endpoint.getType()) {
      case REST:
        String dataString = ProtoBufUtils.toJson(outputsList);
        return queryREST("aggregate", dataString, state, endpoint, true);

      case GRPC:
        if (state.type == PredictiveUnitType.UNKNOWN_TYPE) {
        	try {
        		GenericBlockingStub stub =
        				GenericGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return stub.aggregate(outputsList);
        	} catch (Exception e) {
                logger.error("grpc exception genericStub aggregate", e);
                throw e;
              }
        } else {
        	try {
        		CombinerBlockingStub stub =
        				CombinerGrpc.newBlockingStub(grpcChannelHandler.get(endpoint))
        				.withDeadlineAfter(grpcReadTimeout, TimeUnit.MILLISECONDS)
        				.withMaxInboundMessageSize(grpcMaxMessageSize)
        				.withMaxOutboundMessageSize(grpcMaxMessageSize);
        		return stub.aggregate(outputsList);
        	} catch (Exception e) {
                logger.error("grpc exception combinerStub aggregate", e);
                throw e;
              }
        }
    }
    throw new APIException(
        APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, "no service available");
  }

  private boolean isDefaultData(SeldonMessage message) {
    if (message.getDataOneofCase() == DataOneofCase.DATA) {
      return true;
    }
    return false;
  }

  public static String getUriKey(Endpoint endpoint, String path) {
    StringBuilder sb = new StringBuilder();
    return sb.append(endpoint.getServiceHost())
        .append(":")
        .append(endpoint.getServicePort())
        .append(path)
        .toString();
  }

  private SeldonMessage queryREST(
      String path,
      String dataString,
      PredictiveUnitState state,
      Endpoint endpoint,
      boolean isDefault) {
    long timeNow = System.currentTimeMillis();
    URI uri;
    try {
      final String uriKey = getUriKey(endpoint, path);
      if (uriCache.containsKey(uriKey)) {
        uri = uriCache.get(uriKey);
      } else {
        URIBuilder builder =
            new URIBuilder()
                .setScheme("http")
                .setHost(endpoint.getServiceHost())
                .setPort(endpoint.getServicePort())
                .setPath("/" + path);
        uri = builder.build();
        uriCache.put(uriKey, uri);
      }
    } catch (URISyntaxException e) {
      throw new APIException(
          APIException.ApiExceptionType.ENGINE_INVALID_ENDPOINT_URL,
          "Host: " + endpoint.getServiceHost() + " port:" + endpoint.getServicePort());
    }

    for (int i = 0; i < restRetries; i++) {
      try {
        HttpHeaders headers;
        if (headersCache.containsKey(state.name)) {
          headers = headersCache.get(state.name);
        } else {
          headers = new HttpHeaders();
          headers.setContentType(MediaType.APPLICATION_FORM_URLENCODED);
          headers.add(MODEL_NAME_HEADER, state.name);
          headers.add(MODEL_IMAGE_HEADER, state.imageName);
          headers.add(MODEL_VERSION_HEADER, state.imageVersion);
          headersCache.put(state.name, headers);
        }

        MultiValueMap<String, String> map = new LinkedMultiValueMap<String, String>();
        map.add("json", dataString);
        map.add("isDefault", Boolean.toString(isDefault));

        HttpEntity<MultiValueMap<String, String>> request =
            new HttpEntity<MultiValueMap<String, String>>(map, headers);

        if (logger.isDebugEnabled()) {
          logger.debug("Requesting {}", uri.toString());
        }
        ResponseEntity<String> httpResponse =
            restTemplate.postForEntity(uri, request, String.class);
        try {
          String response = httpResponse.getBody();
          logger.debug(response);
          SeldonMessage.Builder builder = SeldonMessage.newBuilder();
          JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
          return builder.build();
        } finally {
          if (logger.isDebugEnabled()) {
            logger.debug(
                "External prediction server took " + (System.currentTimeMillis() - timeNow) + "ms");
          }
        }
      } catch (ResourceAccessException e) {
        logger.warn("Caught resource access exception ", e);
      } catch (InvalidProtocolBufferException e) {
          logger.error("Invalid protocol buffer during Json Format merge - ", e);
          throw new APIException(
                  APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, e.toString());
      } catch (HttpStatusCodeException e)
      {
          logger.error(
                  "Couldn't retrieve prediction from external prediction server -- bad http return code: "
                          + e.getRawStatusCode());
          handleHttpStatusCodeError(e);
      } catch (Exception e)
      {
          logger.error("Couldn't retrieve prediction from external prediction server - ", e);
          throw new APIException(
                  APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, e.toString());
      }
    }
    logger.error("Failed to retrueve predictions after {} attempts", restRetries);
    throw new APIException(
        APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,
        String.format("Failed to retrieve predictions after %d attempts", restRetries));
  }

  private void handleHttpStatusCodeError(HttpStatusCodeException exception) {
      String response = exception.getResponseBodyAsString();
      SeldonMessage.Builder builder = SeldonMessage.newBuilder();
      try {
          JsonFormat.parser().ignoringUnknownFields().merge(response, builder);
          SeldonMessage seldonMessage = builder.build();
          Status seldonMessageStatus = seldonMessage.getStatus();
          if (seldonMessageStatus == null) {
              throw new APIException(
                      APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR,
                      String.format("Bad return code %d", e.getRawStatusCode()));
          }
          else
          {
              throw new APIException(seldonMessageStatus.getCode(),
                      seldonMessageStatus.getReason(),
                      200,
                      seldonMessageStatus.getInfo());
          }
      } catch (InvalidProtocolBufferException ex)
      {
          logger.error("Invalid protocol buffer during Json Format merge - ", ex);
          throw new APIException(
                  APIException.ApiExceptionType.ENGINE_MICROSERVICE_ERROR, ex.toString());
      }
  }
}
