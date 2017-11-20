package io.seldon.clustermanager.pb;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.util.JsonFormat;
import com.google.protobuf.util.JsonFormat.Printer;


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
        String json = null;
        // json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(message);
        // json =
        // JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().omittingInsignificantWhitespace().print(message);

        Printer jsonPrinter = JsonFormat.printer().preservingProtoFieldNames();
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
        JsonFormat.parser().ignoringUnknownFields().merge(json, messageBuilder);

    }

}
