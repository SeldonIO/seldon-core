package io.seldon.clustermanager.pb;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.util.JsonFormat;

public class ProtoBufUtils {

    public static String toJson(Message message) throws InvalidProtocolBufferException {
        String json = null;
        json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(message);
        return json;
    }

}
