/*
 * Seldon -- open source prediction engine
 * =======================================
 *
 * Copyright 2011-2017 Seldon Technologies Ltd and Rummble Ltd (http://www.seldon.io/)
 *
 * ********************************************************************************************
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * ********************************************************************************************
 */

package io.seldon.engine.exception;

public class APIException extends RuntimeException {

  public enum ApiExceptionType {
    ENGINE_INVALID_JSON(201, "Invalid JSON", 400),
    ENGINE_INVALID_RESPONSE_JSON(201, "Invalid Response JSON", 500),
    ENGINE_INVALID_ENDPOINT_URL(202, "Invalid Endpoint URL", 500),
    ENGINE_MICROSERVICE_ERROR(203, "Microservice error", 500),
    ENGINE_INVALID_ABTEST(204, "Error happened in AB Test Routing", 500),
    ENGINE_INVALID_COMBINER_RESPONSE(204, "Invalid number of predictions from combiner", 500),
    ENGINE_INTERRUPTED(205, "API call interrupted", 500),
    ENGINE_EXECUTION_FAILURE(206, "Execution failure", 500),
    ENGINE_INVALID_ROUTING(207, "Invalid Routing", 500),
    REQUEST_IO_EXCEPTION(208, "IO Exception", 500);

    int id;
    String message;
    int httpCode;

    ApiExceptionType(int id, String message, int httpCode) {
      this.id = id;
      this.message = message;
      this.httpCode = httpCode;
    }
  };

  ApiExceptionType apiExceptionType;
  int id;
  String message;
  int httpCode;
  String info;

  public APIException(ApiExceptionType apiExceptionType, String info) {
    super();
    this.apiExceptionType = apiExceptionType;
    this.info = info;
  }

  public APIException(int id, String message, int httpCode, String info) {
    super();
    this.id = id;
    this.message = message;
    this.httpCode = httpCode;
    this.info = info;
  }

  public String getInfo() {
    return info;
  }

  public void setInfo(String info) {
    this.info = info;
  }

  public int getId() {
    return id;
  }

  public String getMessage()
  {
    return message;
  }
  public int getHttpCode()
  {
    return httpCode;
  }
}
