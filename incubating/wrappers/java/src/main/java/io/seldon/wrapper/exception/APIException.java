/*
 * Seldon -- prediction engine
 * =======================================
 *
 * Copyright 2011-2017 Seldon Technologies Ltd and Rummble Ltd (http://www.seldon.io/)
 *
 * ********************************************************************************************
 *
 * 
 * Copyright (c) 2024 Seldon Technologies Ltd.
 *
 * Use of this software is governed BY
 * (1) the license included in the LICENSE file or
 * (2) if the license included in the LICENSE file is the Business Source License 1.1,
 * the Change License after the Change Date as each is defined in accordance with the LICENSE file.
 *
 *
 * ********************************************************************************************
 */

package io.seldon.wrapper.exception;

/**
 * API Exceptions
 *
 * @author clive
 */
public class APIException extends RuntimeException {

  public enum ApiExceptionType {
    WRAPPER_INVALID_MESSAGE(201, "Invalid prediction message", 500);

    int id;
    String message;
    int httpCode;

    ApiExceptionType(int id, String message, int httpCode) {
      this.id = id;
      this.message = message;
      this.httpCode = httpCode;
    }

    public int getId() {
      return id;
    }

    public String getMessage() {
      return message;
    }

    public int getHttpCode() {
      return httpCode;
    }
  };

  ApiExceptionType apiExceptionType;
  String info;

  public APIException(ApiExceptionType apiExceptionType, String info) {
    super();
    this.apiExceptionType = apiExceptionType;
    this.info = info;
  }

  public ApiExceptionType getApiExceptionType() {
    return apiExceptionType;
  }

  public void setApiExceptionType(ApiExceptionType apiExceptionType) {
    this.apiExceptionType = apiExceptionType;
  }

  public String getInfo() {
    return info;
  }

  public void setInfo(String info) {
    this.info = info;
  }
}
