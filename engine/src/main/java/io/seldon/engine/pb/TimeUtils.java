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
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.Message.Builder;
import com.google.protobuf.MessageOrBuilder;
import com.google.protobuf.Timestamp;
import com.google.protobuf.util.Timestamps;
import io.seldon.engine.pb.JsonFormat.TypeConverter;
import io.seldon.engine.pb.JsonFormat.TypeParser;
import java.io.IOException;
import java.text.ParseException;

public class TimeUtils {

  public static class TimeConverter implements TypeConverter {
    private ByteString toByteString(MessageOrBuilder message) {
      if (message instanceof Message) {
        return ((Message) message).toByteString();
      } else {
        return ((Message.Builder) message).build().toByteString();
      }
    }

    @Override
    public String convert(MessageOrBuilder message) throws IOException {
      Timestamp value = Timestamp.parseFrom(toByteString(message));
      return ("\"" + Timestamps.toString(value) + "\"");
    }
  }

  public static class TimeParser implements TypeParser {

    @Override
    public void merge(JsonElement json, Builder builder) throws InvalidProtocolBufferException {
      try {
        Timestamp value = Timestamps.parse(json.getAsString());
        builder.mergeFrom(value.toByteString());
      } catch (ParseException e) {
        throw new InvalidProtocolBufferException("Failed to parse timestamp: " + json);
      }
    }
  }
}
