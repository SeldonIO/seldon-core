package io.seldon.apife;

import java.io.IOException;
import java.nio.charset.Charset;
import java.nio.file.Files;
import java.nio.file.Paths;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;

import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.Time;
import io.kubernetes.client.proto.Meta.Timestamp;
import io.kubernetes.client.proto.Resource.Quantity;
import io.seldon.apife.pb.IntOrStringUtils;
import io.seldon.apife.pb.JsonFormat;
import io.seldon.apife.pb.QuantityUtils;
import io.seldon.apife.pb.TimeUtils;

public class SeldonTestBase {
	protected String readFile(String path, Charset encoding) 
			  throws IOException 
	 {
		 byte[] encoded = Files.readAllBytes(Paths.get(path));
		 return new String(encoded, encoding);
	 }	
	
	protected <T extends Message.Builder> void updateMessageBuilderFromJson(T messageBuilder, String json) throws InvalidProtocolBufferException {
    JsonFormat.parser().ignoringUnknownFields()
    .usingTypeParser(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringParser())
    .usingTypeParser(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityParser())
    .usingTypeParser(Time.getDescriptor().getFullName(), new TimeUtils.TimeParser())
    .usingTypeParser(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeParser()) 
    .merge(json, messageBuilder);
}
}
