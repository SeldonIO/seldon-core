package io.seldon.clustermanager.pb;

import org.junit.Test;

import com.google.protobuf.InvalidProtocolBufferException;

import io.kubernetes.client.proto.Meta.Time;

public class TimeFormatTest {

    @Test
    public void parseTimestamp() throws InvalidProtocolBufferException
    {
        String json = "\"2017-11-23T20:37:27Z\"";
        Time.Builder builder = Time.newBuilder();
        JsonFormat.parser().usingTypeParser(Time.getDescriptor().getFullName(),new TimeUtils.TimeParser()).merge(json, builder);
    }
    
}
