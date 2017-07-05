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
	public static final int AUTHENTICATION_REQ = 1;
	public static final int NOT_SSL_CONN = 2;
	public static final int NOT_AUTHORIZED_CONS = 3;
	public static final int NOT_VALID_TOKEN = 4;
	public static final int NOT_VALID_KEY_CONS = 5;
	public static final int NOT_VALID_SECRET_CONS = 6;
	public static final int NOT_VALID_TOKEN_KEY = 7;
	public static final int NOT_VALID_TOKEN_EXPIRED = 8;
	public static final int NOT_VALID_CONNECTION = 9;
	public static final int JSON_ERROR = 10;
	public static final int HTTPMETHOD_NOT_VALID = 11;
	public static final int NOT_SECURE_TOKEN = 12;
	public static final int GENERIC_ERROR = 13;
	public static final int INTERNAL_DB_ERROR = 14;
	public static final int NOT_SPECIFIED_TOKEN = 15;
	public static final int NOT_SPECIFIED_CONS_KEY = 16;
	public static final int NOT_SPECIFIED_CONS_SECRET = 17;
	public static final int USER_ID = 18;
	public static final int ITEM_ID = 19;
	public static final int METHOD_NOT_AUTHORIZED = 20;
	public static final int RESOURCE_NOT_FOUND = 21;
	public static final int NUMBER_FORMAT_NOT_VALID = 22;
	public static final int USER_NOT_FOUND = 23;
	public static final int ITEM_NOT_FOUND = 24;
	public static final int SERVICE_NOT_IMPLEMENTED = 25;
	public static final int INCORRECT_FIELD = 26;
    public static final int ITEM_TYPE_NOT_FOUND = 27;
    public static final int ACTION_TYPE_NOT_FOUND = 28;
	public static final int ITEM_DUPLICATED = 29;
	public static final int USER_DUPLICATED = 30;
	public static final int CANNOT_CLONE_CFALGORITHM = 31;
    public static final int CANT_LOAD_CONTENT_MODEL = 32;
    public static final int INCOMPLETE_ATTRIBUTE_ADDITION = 33;
    public static final int CONCURRENT_ITEM_UPDATE = 34;
    public static final int CONCURRENT_USER_UPDATE = 35;
    public static final int FACEBOOK_RESPONSE = 36;
	public static final int NOT_VALID_STRATEGY = 37;
	public static final int INVALID_JSON = 38;
	public static final int PLUGIN_NOT_ENABLED = 39;
	public static final int DEPLOYMENT_NOT_FOUND = 40;
	public static final int COMBINER_ERROR = 41;
	public static final int MICROSERVICE_ERROR = 42;
	public static final int ABTEST_ERROR = 43;

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
			case AUTHENTICATION_REQ:
				error_msg = "Authentication Required";
				httpResponse =  400;
				break;
			case NOT_SSL_CONN:
				error_msg = "Secure connection required";
				httpResponse = 400;
				break;
			case NOT_AUTHORIZED_CONS:
				error_msg = "Consumer not Authorized";
				httpResponse = 400;
				break;
			case NOT_VALID_TOKEN:
				error_msg = "Token no longer Valid";
				httpResponse = 400;
				break;
			case NOT_VALID_KEY_CONS:
				error_msg = "Consumer key not recognized";
				httpResponse = 400;
				break;
			case NOT_VALID_SECRET_CONS:
				error_msg = "Consumer secret not correct";
				httpResponse = 400;
				break;
			case NOT_VALID_TOKEN_KEY:
				error_msg = "Token value not recognized";
				httpResponse = 400;
				break;
			case NOT_VALID_TOKEN_EXPIRED:
				error_msg = "Token expired";
				httpResponse = 400;
				break;
			case NOT_VALID_CONNECTION:
				error_msg = "Connection problem";
				httpResponse = 400;
				break;
			case JSON_ERROR:
				error_msg = "Error in generating json object";
				httpResponse = 400;
				break;
			case HTTPMETHOD_NOT_VALID:
				error_msg = "Used Http Method is not supported for the required action";
				httpResponse = 400;
				break;
			case NOT_SECURE_TOKEN:
				error_msg = "Token sent in a not secure way";
				httpResponse = 400;
				break;
			case GENERIC_ERROR:
				error_msg = "A generic error occurred";
				httpResponse = 400;
				break;
			case INTERNAL_DB_ERROR:
				error_msg = "Internal Error : Database problem";
				httpResponse = 400;
				break;
			case NOT_SPECIFIED_TOKEN:
				error_msg = "To access the requested resource a valid token must be specified";
				httpResponse = 400;
				break;
			case NOT_SPECIFIED_CONS_KEY:
				error_msg = "Please specify your consumer_key";
				httpResponse = 400;
				break;
			case NOT_SPECIFIED_CONS_SECRET:
				error_msg = "Please specify your consumer_secret";
				httpResponse = 400;
				break;
			case USER_ID:
				error_msg = "The format of the field id for the user is not valid";
				httpResponse = 400;
				break;
			case ITEM_ID:
				error_msg = "The format of the field id for the item is not valid";
				httpResponse = 400;
				break;
			case METHOD_NOT_AUTHORIZED:
				error_msg = "The token does not authorized to the requested method";
				httpResponse = 400;
				break;
			case RESOURCE_NOT_FOUND:
				error_msg = "Requested resource not found";
				httpResponse = 400;
				break;
			case NUMBER_FORMAT_NOT_VALID:
				error_msg = "Parameter format not valid: a number is required";
				httpResponse = 400;
				break;
			case USER_NOT_FOUND:
				error_msg = "Requested user not found";
				httpResponse = 400;
				break;
			case ITEM_NOT_FOUND:
				error_msg = "Requested item not found";
				httpResponse = 400;
				break;
			case SERVICE_NOT_IMPLEMENTED:
				error_msg = "service not implemented";
				httpResponse = 400;
				break;
			case INCORRECT_FIELD:
				error_msg = "The supplied fields are incorrect or incompatible with your current integration. Please contact Rummble Labs support.";
				httpResponse = 400;
				break;
			case ITEM_DUPLICATED:
				error_msg = "The item is already in the system. Use PUT to update the item";
				httpResponse = 400;
				break;
			case USER_DUPLICATED:
				error_msg = "The user is already in the system. Use PUT to update the user";
				httpResponse = 400;
				break;
			case CANT_LOAD_CONTENT_MODEL:
				error_msg = "Can't load content model for this client. Please contact Rummble Labs support";
				httpResponse = 400;
				break;
            case ITEM_TYPE_NOT_FOUND:
                error_msg = "Item type not found.";
                httpResponse = 400;
                break;
            case INCOMPLETE_ATTRIBUTE_ADDITION:
                error_msg = "Incomplete attribute addition (see attribute_failures map).";
                httpResponse = 400;
                break;
            case CONCURRENT_ITEM_UPDATE:
                error_msg = "Concurrent item update; please retrieve the item, compare and retry.";
                httpResponse = 400;
                break;
            case CONCURRENT_USER_UPDATE:
                error_msg = "Concurrent user update; please retrieve the user, compare and retry.";
                httpResponse = 400;
                break;
            case FACEBOOK_RESPONSE:
                error_msg = "Facebook response status exception, not code 1. https://developers.facebook.com/docs/reference/api/errors/";
                httpResponse = 500;
                break;
			case NOT_VALID_STRATEGY:
				error_msg = "Invalid or Null Strategy. The default strategy or a per client strategy needs to be set correctly.";
				httpResponse = 500;
				break;
			case INVALID_JSON:
				error_msg = "Invalid JSON";
				httpResponse = 400;
				break;
			case DEPLOYMENT_NOT_FOUND:
				error_msg = "Deployment not found";
				httpResponse = 400;
				break;
			case COMBINER_ERROR:
				error_msg = "Model outputs cannot be combined";
				httpResponse = 500;
				break;
			case MICROSERVICE_ERROR:
				error_msg = "Couldn't retrieve prediction from external microservice";
				httpResponse = 500;
				break;
			case ABTEST_ERROR:
				error_msg = "An error occured in the AB Test unit";
				httpResponse = 500;
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
