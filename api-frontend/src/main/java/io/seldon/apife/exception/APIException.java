/*
 * Seldon -- open source prediction engine
 * =======================================
 *
 * Copyright 2011-2015 Seldon Technologies Ltd and Rummble Ltd (http://www.seldon.io/)
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

import java.util.Map;

/**
 * @author claudio
 */

public class APIException extends RuntimeException {
	//possible errors
	static final String ERROR = "error";

	public static final int INVALID_JSON = 101;
	public static final int INVALID_ENDPOINT_URL = 102;	
	public static final int MICROSERVICE_ERROR = 103;	
	public static final int NO_RUNNING_DEPLOYMENT = 104;	

	//ATTRIBUTES
    int error_id;
    String error_msg;
    int httpResponse;

    private Map<String, String> failureMap;

    //CONSTRUCTOR
	public APIException(int error_id) {
		super();
		this.error_id = error_id;
		switch(error_id) {

		case INVALID_JSON:
			error_msg = "Invalid JSON";
			httpResponse = 400;
			break;
		case INVALID_ENDPOINT_URL:
			error_msg = "Invalid Endpoint URL";
			httpResponse = 400;
			break;
		case MICROSERVICE_ERROR:
			error_msg = "Microservice error";
			httpResponse = 400;
			break;
		case NO_RUNNING_DEPLOYMENT:
			error_msg = "No Running Deployment";
			httpResponse = 400;
			break;

		}
	}
	
	//GETTER AND SETTER
	public String toString() {
		return error_id + " : " + error_msg;
	}

	public int getError_id() {
		return error_id;
	}

	public void setError_id(int errorId) {
		error_id = errorId;
	}

	public String getError_msg() {
		return error_msg;
	}

	public void setError_msg(String errorMsg) {
		error_msg = errorMsg;
	}

	public int getHttpResponse() {
		return httpResponse;
	}

	public void setHttpResponse(int httpResponse) {
		this.httpResponse = httpResponse;
	}

    public Map<String, String> getFailureMap() {
        return failureMap;
    }

    public void setFailureMap(Map<String, String> failureMap) {
        this.failureMap = failureMap;
    }
}
