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
package io.seldon.engine;

import com.google.protobuf.InvalidProtocolBufferException;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.protos.PredictionProtos.Status;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;

@ControllerAdvice
public class ExceptionControllerAdvice {

  @ExceptionHandler(APIException.class)
  public ResponseEntity<String> handleUnauthorizedException(APIException exception)
      throws InvalidProtocolBufferException {

    Status.Builder statusBuilder = Status.newBuilder();
    statusBuilder.setCode(exception.getApiExceptionType().getId());
    statusBuilder.setReason(exception.getApiExceptionType().getMessage());
    statusBuilder.setInfo(exception.getInfo());
    statusBuilder.setStatus(Status.StatusFlag.FAILURE);

    Status status = statusBuilder.build();
    String json;
    json = ProtoBufUtils.toJson(status);
    return new ResponseEntity<String>(
        json, HttpStatus.valueOf(exception.getApiExceptionType().getHttpCode()));
  }
}
