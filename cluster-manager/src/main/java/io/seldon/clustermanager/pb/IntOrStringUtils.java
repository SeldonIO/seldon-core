package io.seldon.clustermanager.pb;

import java.io.IOException;

import com.google.gson.JsonElement;
import com.google.gson.JsonPrimitive;
import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.MessageOrBuilder;
import com.google.protobuf.Message.Builder;

import io.kubernetes.client.proto.IntStr.IntOrString;
import io.seldon.clustermanager.pb.JsonFormat.TypeConverter;
import io.seldon.clustermanager.pb.JsonFormat.TypeParser;

public class IntOrStringUtils {

	public static class IntOrStringConverter implements TypeConverter
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
			IntOrString is = IntOrString.parseFrom(toByteString(message));
			if (is.hasStrVal())
				return "\"" + is.getStrVal() + "\"";
			else
				return ""+is.getIntVal();
			
		}
		
	}
	
	public static class IntOrStringParser implements TypeParser {
		@Override
		public void merge(JsonElement json, Builder builder) throws InvalidProtocolBufferException {
			if (json instanceof JsonPrimitive) {
		        JsonPrimitive primitive = (JsonPrimitive) json;
		        if (primitive.isString())
		        {
		        	IntOrString.Builder b = IntOrString.newBuilder().setStrVal(primitive.getAsString());
		        	builder.mergeFrom(b.build().toByteArray());
		        }
		        else if (primitive.isNumber())
		        {
		        	IntOrString.Builder b = IntOrString.newBuilder().setIntVal(primitive.getAsInt());
		        	builder.mergeFrom(b.build().toByteArray());
		        }	
		        }
			else throw new InvalidProtocolBufferException("Can't decode io.kubernetes.client.proto.IntOrSting from "+json.toString());
		}
	}
}
