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

package io.seldon.apife.exception;

public class SeldonAPIException extends RuntimeException {

	public enum ApiExceptionType { 
		
		APIFE_INVALID_JSON(101,"Invalid JSON",400),
		APIFE_INVALID_ENDPOINT_URL(102,"Invalid Endpoint URL",500),	
		APIFE_MICROSERVICE_ERROR(103,"Microservice error",500),
		APIFE_NO_RUNNING_DEPLOYMENT(104,"No Running Deployment",500),
		APIFE_INVALID_RESPONSE_JSON(105,"Invalid Response JSON",400),
		APIFE_GRPC_NO_PRINCIPAL_FOUND(105,"No OAuth principal found",400),
	    APIFE_GRPC_NO_GRPC_CHANNEL_FOUND(106,"No Managed Channel found",400);
		
		int id;
		String message;
		int httpCode;

		ApiExceptionType(int id,String message,int httpCode) {
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

   public SeldonAPIException(ApiExceptionType apiExceptionType,String info) {
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
