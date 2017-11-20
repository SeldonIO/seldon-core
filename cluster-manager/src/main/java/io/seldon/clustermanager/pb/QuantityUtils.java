package io.seldon.clustermanager.pb;

import java.io.IOException;

import com.google.gson.JsonElement;
import com.google.gson.JsonPrimitive;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.Message.Builder;
import com.google.protobuf.MessageOrBuilder;

import io.kubernetes.client.proto.Resource.Quantity;
import io.seldon.clustermanager.pb.JsonFormat.TypeConverter;
import io.seldon.clustermanager.pb.JsonFormat.TypeParser;

public class QuantityUtils {

	public static class QuantityConverter implements TypeConverter
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
			Quantity q = Quantity.parseFrom(toByteString(message));
			return "\"" + q.getString() + "\"";
		}
		
	}
	
	public static class QuantityParser implements TypeParser {

		@Override
		public void merge(JsonElement json, Builder builder) throws InvalidProtocolBufferException {
			if (json instanceof JsonPrimitive) {
		        JsonPrimitive primitive = (JsonPrimitive) json;
		        if (primitive.isString())
		        {
		        	Quantity.Builder b = Quantity.newBuilder().setString(primitive.getAsString());
		        	builder.mergeFrom(b.build().toByteArray());
		        }
		        else throw new InvalidProtocolBufferException("Can't decode io.kubernetes.client.proto.resource.Quantity from "+json.toString());
			}
		}
		
	}
}
