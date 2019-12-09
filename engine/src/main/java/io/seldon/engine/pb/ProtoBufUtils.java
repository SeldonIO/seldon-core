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
package io.seldon.engine.pb;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.Time;
import io.kubernetes.client.proto.Meta.Timestamp;
import io.kubernetes.client.proto.Resource.Quantity;
import io.seldon.engine.pb.JsonFormat.Printer;

public class ProtoBufUtils {

  private ProtoBufUtils() {}

  /**
   * Serialize a protobuf message to JSON, indicating if whitespace needs to be stripped.
   *
   * @param message The protobuf message to serialize.
   * @param omittingInsignificantWhitespace True if needs to be stripped of whitespace.
   * @return json string
   * @throws InvalidProtocolBufferException
   */
  public static String toJson(Message message, boolean omittingInsignificantWhitespace)
      throws InvalidProtocolBufferException {
    Printer jsonPrinter =
        JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames();
    if (omittingInsignificantWhitespace) {
      jsonPrinter = jsonPrinter.omittingInsignificantWhitespace();
    }
    return jsonPrinter.print(message);
  }

  /**
   * Serialize a protobuf message to JSON, allowing the inclusion of whitespace.
   *
   * @param message The protobuf message to serialize.
   * @return json string
   * @throws InvalidProtocolBufferException
   */
  public static String toJson(Message message) throws InvalidProtocolBufferException {
    boolean omittingInsignificantWhitespace = false;
    return toJson(message, omittingInsignificantWhitespace);
  }

  public static <T extends Message.Builder> void updateMessageBuilderFromJson(
      T messageBuilder, String json) throws InvalidProtocolBufferException {
    JsonFormat.parser()
        .ignoringUnknownFields()
        .usingTypeParser(
            IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringParser())
        .usingTypeParser(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityParser())
        .usingTypeParser(Time.getDescriptor().getFullName(), new TimeUtils.TimeParser())
        .usingTypeParser(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeParser())
        .merge(json, messageBuilder);
  }
}
