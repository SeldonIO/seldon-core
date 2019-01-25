/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.clustermanager.k8s;

public class Constants {
	public static final String LABEL_SELDON_ID = "seldon-deployment-id";
	public static final String LABEL_SELDON_SVCORCH = "seldon-deployment-contains-svcorch";
    public static final String STATE_CREATING = "Creating";
    public static final String STATE_FAILED = "Failed";
    public static final String STATE_AVAILABLE = "Available";
    
    public static final String ENGINE_JAVA_OPTS_ANNOTATION = "seldon.io/engine-java-opts";
    public static final String ENGINE_SEPARATE_ANNOTATION = "seldon.io/engine-separate-pod";
    public static final String REST_READ_TIMEOUT_ANNOTATION = "seldon.io/rest-read-timeout";
    public static final String GRPC_READ_TIMEOUT_ANNOTATION = "seldon.io/grpc-read-timeout";
    public static final String HEADLESS_SVC_ANNOTATION = "seldon.io/headless-svc";
    
    /* Ambassador annotations */
    public static final String AMBASSADOR_WEIGHT_ANNOTATION = "seldon.io/ambassador-weight";
    public static final String AMBASSADOR_SERVICE_ANNOTATION = "seldon.io/ambassador-service-name";    
    public static final String AMBASSADOR_SHADOW_ANNOTATION = "seldon.io/ambassador-shadow";    
    public static final String AMBASSADOR_HEADER_ANNOTATION = "seldon.io/ambassador-header";        
    public static final String AMBASSADOR_REGEX_HEADER_ANNOTATION = "seldon.io/ambassador-regex-header";        
}
