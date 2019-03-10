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

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.Time;
import io.kubernetes.client.proto.Meta.Timestamp;
import io.kubernetes.client.proto.Resource.Quantity;
import io.seldon.clustermanager.pb.IntOrStringUtils;
import io.seldon.clustermanager.pb.JsonFormat;
import io.seldon.clustermanager.pb.JsonFormat.Printer;
import io.seldon.clustermanager.pb.QuantityUtils;
import io.seldon.clustermanager.pb.TimeUtils;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

/**
 * Set of general utility methods
 * @author clive
 *
 */
public class SeldonDeploymentUtils {
	private  static Logger logger = LoggerFactory.getLogger(SeldonDeploymentUtils.class.getName());
	
	/**
	 * Parse some JSON into a Seldon Deployment. Will throw exception if unknown fields are found.
	 * @param json
	 * @return SeldonDeployment
	 * @throws InvalidProtocolBufferException
	 */
	public static SeldonDeployment jsonToSeldonDeployment(String json) throws InvalidProtocolBufferException {
		SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder();
		JsonFormat.parser()//.ignoringUnknownFields()
			.usingTypeParser(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringParser())
			.usingTypeParser(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityParser())
            .usingTypeParser(Time.getDescriptor().getFullName(), new TimeUtils.TimeParser())
            .usingTypeParser(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeParser())            
			.merge(json, mlBuilder);
		return mlBuilder.build();
	}
	
	/**
	 * Turn a SeldonDeployment into a JSON string.
	 * @param mlDep
	 * @param omittingWhitespace
	 * @return
	 * @throws InvalidProtocolBufferException
	 */
	public static String toJson(SeldonDeployment mlDep,boolean omittingWhitespace) throws InvalidProtocolBufferException
	{
		Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames()
				.usingTypeConverter(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringConverter())
				.usingTypeConverter(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityConverter())
				.usingTypeConverter(Time.getDescriptor().getFullName(), new TimeUtils.TimeConverter())
                .usingTypeConverter(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeConverter());
		if (omittingWhitespace)
		    jsonPrinter = jsonPrinter.omittingInsignificantWhitespace();
		return jsonPrinter.print(mlDep);
				
	}
	
	/**
	 * Get the namespace from a Seldon Deployment choosing default if none found
	 * @param d
	 * @return
	 */
	public static String getNamespace(SeldonDeployment d)
	{
	    if (StringUtils.isEmpty(d.getMetadata().getNamespace()))
	        return "default";
	    else
	        return d.getMetadata().getNamespace();
	}
	
	/**
	 * Get the version from the version field of a resource. Assumes a two part string separated by slash
	 * @param apiVersion
	 * @return The version part - e.g. v1alpha2
	 */
	public static String getVersionFromApiVersion(String apiVersion)
	{
		return apiVersion.split("/")[1];
	}
	
	public static boolean hasSeparateEnginePodAnnotation(SeldonDeployment mlDep)
	{
		return Boolean.parseBoolean(mlDep.getSpec().getAnnotationsOrDefault(Constants.ENGINE_SEPARATE_ANNOTATION, "false"));
	}
	
	public static boolean hasHeadlessSvcAnnotation(SeldonDeployment mlDep)
	{
		return Boolean.parseBoolean(mlDep.getSpec().getAnnotationsOrDefault(Constants.HEADLESS_SVC_ANNOTATION, "false"));
	}
	
}
