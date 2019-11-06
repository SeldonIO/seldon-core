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

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.ListValue;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import java.util.Arrays;
import java.util.Map;
import org.junit.Assert;
import org.junit.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TestPredictionProto {
  private static Logger log = LoggerFactory.getLogger(TestPredictionProto.class);

  @Test
  public void testParseJsonExtraFields() throws InvalidProtocolBufferException {
    String json = "{\"x\":1.0,\"request\":{\"values\":[[1.0],[2.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    // TODO: Nothing gets serialised back. Is this alright?
    Assert.assertEquals(0, request.getSerializedSize());
  }

  @Test
  public void testParseCustomJson() throws InvalidProtocolBufferException {
    String json = "{\"data\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    ListValue ndarray = request.getData().getNdarray();
    Assert.assertEquals(2, ndarray.getValuesCount());

    ListValue ndarray1 = ndarray.getValues(1).getListValue();
    Assert.assertEquals(3.0, ndarray1.getValues(0).getNumberValue(), 0.01);
  }

  @Test
  public void testParseTagsArray() throws InvalidProtocolBufferException {
    String json =
        "{\"meta\":{\"tags\":{\"user\":[\"a\",\"b\"],\"gender\":\"female\"}},\"data\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    Map<String, Value> metaTags = request.getMeta().getTags();
    Assert.assertEquals(2, metaTags.size());

    Value gender = metaTags.get("gender");
    Assert.assertEquals("female", gender.getStringValue());

    ListValue user = metaTags.get("user").getListValue();
    Assert.assertEquals("a", user.getValues(0).getStringValue());

    ListValue ndarray = request.getData().getNdarray();
    Assert.assertEquals(2, ndarray.getValuesCount());

    ListValue ndarray0 = ndarray.getValues(0).getListValue();
    Assert.assertEquals(2.0, ndarray0.getValues(1).getNumberValue(), 0.01);
  }

  @Test
  public void testParseJson() throws InvalidProtocolBufferException {
    String json = "{\"request\":{\"values\":[[1.0],[2.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    // TODO: Nothing gets serialised back. Is this alright?
    Assert.assertEquals(0, request.getSerializedSize());
  }

  @Test
  public void testParseJsonData() throws InvalidProtocolBufferException {
    String json = "{\"jsonData\":{\"key1\":\"bar\",\"key2\": 23,\"key3\":2.3}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    Struct jsonData = request.getJsonData().getStructValue();
    Assert.assertEquals(3, jsonData.getFieldsCount());

    Value key1 = jsonData.getFieldsOrThrow("key1");
    Assert.assertEquals("key1", key1.getStringValue());

    // TODO: How to read an Int32Value from a Struct
    Value key2 = jsonData.getFieldsOrThrow("key2");
    Assert.assertEquals(23, key2.getNumberValue());

    Value key3 = jsonData.getFieldsOrThrow("key3");
    Assert.assertEquals(2.3, key3.getNumberValue());
  }

  @Test
  public void testDefaultRequest() throws InvalidProtocolBufferException {
    String[] features = {"a", "b"};
    Double[] values = {1.0, 2.0, 1.5, 2.2};
    DefaultData.Builder defB = DefaultData.newBuilder();
    defB.addAllNames(Arrays.asList(features));
    defB.setTensor(
        Tensor.newBuilder()
            .addShape(1)
            .addShape(values.length)
            .addAllValues(Arrays.asList(values))
            .build());
    SeldonMessage.Builder b = SeldonMessage.newBuilder();
    Value v;
    Value v1 = Value.newBuilder().setNumberValue(1.0).build();

    b.setData(defB.build()).setMeta(Meta.newBuilder().putTags("key", v1).build());
    SeldonMessage request = b.build();
    String json = ProtoBufUtils.toJson(request);

    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
    SeldonMessage request2 = b2.build();
    String json2 = ProtoBufUtils.toJson(request2);

    Assert.assertEquals(json, json2);
  }

  @Test
  public void customBytesRequest() throws InvalidProtocolBufferException {
    String customData = "{\"c\":1.0}";
    SeldonMessage.Builder b = SeldonMessage.newBuilder();
    b.setBinData(ByteString.copyFrom(customData.getBytes()));
    SeldonMessage request = b.build();

    String json = ProtoBufUtils.toJson(request);

    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
    SeldonMessage request2 = b2.build();
    String json2 = ProtoBufUtils.toJson(request2);

    Assert.assertEquals(json, json2);
  }

  @Test
  public void customStringRequest() throws InvalidProtocolBufferException {
    String customData = "{\"c\":1.0}";
    SeldonMessage.Builder b = SeldonMessage.newBuilder();
    b.setStrData(customData);
    SeldonMessage request = b.build();

    String json = ProtoBufUtils.toJson(request);

    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
    SeldonMessage request2 = b2.build();
    String json2 = ProtoBufUtils.toJson(request2);

    Assert.assertEquals(json, json2);
  }
}
