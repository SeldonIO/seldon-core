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
package io.seldon.clustermanager.pb;

import java.lang.reflect.Type;

import com.google.gson.reflect.TypeToken;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;

import io.kubernetes.client.JSON;
import io.kubernetes.client.models.V1PodTemplateSpec;
import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.Time;
import io.kubernetes.client.proto.Meta.Timestamp;
import io.kubernetes.client.proto.Resource.Quantity;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.seldon.clustermanager.k8s.SeldonDeploymentException;
import io.seldon.clustermanager.pb.JsonFormat.Printer;


public class ProtoBufUtils {

    private ProtoBufUtils() {
    }

    /**
     * Serialize a protobuf message to JSON, indicating if whitespace needs to be stripped.
     * 
     * @param message
     *            The protobuf message to serialize.
     * @param omittingInsignificantWhitespace
     *            True if needs to be stripped of whitespace.
     * @return json string
     * @throws InvalidProtocolBufferException
     */
    public static String toJson(Message message, boolean omittingInsignificantWhitespace, boolean defaultingFields) throws InvalidProtocolBufferException {
        Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames()
                .usingTypeConverter(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringConverter())
                .usingTypeConverter(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityConverter())
                .usingTypeConverter(Time.getDescriptor().getFullName(), new TimeUtils.TimeConverter())
                .usingTypeConverter(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeConverter());
        if (omittingInsignificantWhitespace) {
            jsonPrinter = jsonPrinter.omittingInsignificantWhitespace();
        }
        if (defaultingFields)
        	jsonPrinter = jsonPrinter.includingDefaultValueFields();
        return jsonPrinter.print(message);
    }
    
    /**
     * Serialize a protobuf message to JSON, allowing the inclusion of whitespace.
     * 
     * @param message
     *            The protobuf message to serialize.
     * @return json string
     * @throws InvalidProtocolBufferException
     */
    public static String toJson(Message message) throws InvalidProtocolBufferException {
    	boolean omittingInsignificantWhitespace = false;
    	boolean defaultingFields = false;
        return toJson(message, omittingInsignificantWhitespace,defaultingFields);
    }

    public static <T extends Message.Builder> void updateMessageBuilderFromJson(T messageBuilder, String json) throws InvalidProtocolBufferException {
        JsonFormat.parser().ignoringUnknownFields()
            .usingTypeParser(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringParser())
            .usingTypeParser(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityParser())
            .usingTypeParser(Time.getDescriptor().getFullName(), new TimeUtils.TimeParser())
            .usingTypeParser(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeParser())            
        .merge(json, messageBuilder);

    }
    
    public static <T> T convertProtoToModel(Message m,Type type) throws InvalidProtocolBufferException
    {
         JSON json = new JSON();
         return (T) json.deserialize(toJson(m), type);
    }
    
    
    public static V1PodTemplateSpec convertProtoToModel(PodTemplateSpec protoTemplateSpec) throws InvalidProtocolBufferException, SeldonDeploymentException
    {
         Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
         String ptsJson = jsonPrinter.print(protoTemplateSpec);
         JSON json = new JSON();
         Type returnType = new TypeToken<V1PodTemplateSpec>(){}.getType();
         V1PodTemplateSpec podTemplateSpec = (V1PodTemplateSpec) json.deserialize(ptsJson, returnType);
         //return fixProbes(protoTemplateSpec, podTemplateSpec);
         return podTemplateSpec;
    }

}
