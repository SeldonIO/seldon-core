package io.seldon.clustermanager.pb;

import java.io.IOException;
import java.text.ParseException;

import com.google.gson.JsonElement;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.Message.Builder;
import com.google.protobuf.MessageOrBuilder;
import com.google.protobuf.Timestamp;
import com.google.protobuf.util.Timestamps;

import io.seldon.clustermanager.pb.JsonFormat.TypeConverter;
import io.seldon.clustermanager.pb.JsonFormat.TypeParser;

public class TimeUtils {

    public static class TimeConverter implements TypeConverter
    {
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
