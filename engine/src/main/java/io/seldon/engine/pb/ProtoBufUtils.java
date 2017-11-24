package io.seldon.engine.pb;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;

import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.Time;
import io.kubernetes.client.proto.Meta.Timestamp;
import io.kubernetes.client.proto.Resource.Quantity;
import io.seldon.engine.pb.JsonFormat.Printer;

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
    public static String toJson(Message message, boolean omittingInsignificantWhitespace) throws InvalidProtocolBufferException {
        String json = null;
        // json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(message);
        // json =
        // JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().omittingInsignificantWhitespace().print(message);

        Printer jsonPrinter = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames();
        if (omittingInsignificantWhitespace) {
            jsonPrinter = jsonPrinter.omittingInsignificantWhitespace();
        }
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
        return toJson(message, omittingInsignificantWhitespace);
    }

    public static <T extends Message.Builder> void updateMessageBuilderFromJson(T messageBuilder, String json) throws InvalidProtocolBufferException {
    	JsonFormat.parser().ignoringUnknownFields()
    	.usingTypeParser(IntOrString.getDescriptor().getFullName(), new IntOrStringUtils.IntOrStringParser())
        .usingTypeParser(Quantity.getDescriptor().getFullName(), new QuantityUtils.QuantityParser())
        .usingTypeParser(Time.getDescriptor().getFullName(), new TimeUtils.TimeParser())
        .usingTypeParser(Timestamp.getDescriptor().getFullName(), new TimeUtils.TimeParser()) 
    	.merge(json, messageBuilder);
    }

}
