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

import com.google.gson.JsonElement;
import com.google.gson.JsonPrimitive;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.Message.Builder;
import com.google.protobuf.MessageOrBuilder;
import io.kubernetes.client.proto.Resource.Quantity;
import io.seldon.engine.pb.JsonFormat.TypeConverter;
import io.seldon.engine.pb.JsonFormat.TypeParser;
import java.io.IOException;

public class QuantityUtils {

  public static class QuantityConverter implements TypeConverter {
    private ByteString toByteString(MessageOrBuilder message) {
      if (message instanceof Message) {
        return ((Message) message).toByteString();
      } else {
        return ((Message.Builder) message).build().toByteString();
      }
    }

    @Override
    public String convert(MessageOrBuilder message) throws IOException {
      Quantity q = Quantity.parseFrom(toByteString(message));
      return "\"" + q.getString() + "\"";
    }
  }

  public static class QuantityParser implements TypeParser {

    @Override
    public void merge(JsonElement json, Builder builder) throws InvalidProtocolBufferException {
      if (json instanceof JsonPrimitive) {
        JsonPrimitive primitive = (JsonPrimitive) json;
        if (primitive.isString()) {
          Quantity.Builder b = Quantity.newBuilder().setString(primitive.getAsString());
          builder.mergeFrom(b.build().toByteArray());
        } else
          throw new InvalidProtocolBufferException(
              "Can't decode io.kubernetes.client.proto.resource.Quantity from " + json.toString());
      }
    }
  }
}
