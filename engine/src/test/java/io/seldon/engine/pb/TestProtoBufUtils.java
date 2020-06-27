package io.seldon.engine.pb;

import com.google.protobuf.InvalidProtocolBufferException;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import org.junit.Assert;
import org.junit.Test;

public class TestProtoBufUtils {
  private String toJson(SeldonMessage request) throws InvalidProtocolBufferException {
    String withNewlines = ProtoBufUtils.toJson(request, true);
    return withNewlines.replace("\n", "").replace("\r", "");
  }

  @Test
  public void testSeldonMessage() throws InvalidProtocolBufferException {
    String json = "{\"data\":{\"ndarray\":[[1.5,2.0],[3.223,4.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage message = builder.build();

    String serialised = toJson(message);
    String expected = "{\"data\":{\"names\":[],\"ndarray\":[[1.5,2],[3.223,4]]}}";
    Assert.assertEquals(expected, serialised);
  }

  @Test
  public void testSeldonMessageWithInt() throws InvalidProtocolBufferException {
    String json =
        "{\"jsonData\":{\"vocab_size\":275,\"input\":\"my sentence\",\"threshold\":0.45}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage message = builder.build();

    String serialised = toJson(message);
    String expected =
        "{\"jsonData\":{\"vocab_size\":275,\"input\":\"my sentence\",\"threshold\":0.45}}";
    Assert.assertEquals(expected, serialised);
  }
}
