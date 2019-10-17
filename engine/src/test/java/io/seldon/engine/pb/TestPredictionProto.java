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
import com.google.protobuf.Value;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import org.junit.Assert;
import org.junit.Test;

public class TestPredictionProto {

  @Test
  public void parse_json_extra_fields() throws InvalidProtocolBufferException {
    String json = "{\"x\":1.0,\"request\":{\"values\":[[1.0],[2.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    String json2 = ProtoBufUtils.toJson(request);

    System.out.println(json2);
  }

  @Test
  public void parse_custom_json() throws InvalidProtocolBufferException {
    String json = "{\"data\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    Assert.assertEquals(2, request.getData().getNdarray().getValuesCount());

    String json2 = ProtoBufUtils.toJson(request);

    System.out.println(json2);
  }

  @Test
  public void parse_tags_array() throws InvalidProtocolBufferException {
    String json =
        "{\"meta\":{\"tags\":{\"user\":[\"a\",\"b\"],\"gender\":\"female\"}},\"data\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    Assert.assertEquals(2, request.getData().getNdarray().getValuesCount());

    String json2 = ProtoBufUtils.toJson(request);

    System.out.println(json2);
  }

  @Test
  public void parse_json() throws InvalidProtocolBufferException {
    String json = "{\"request\":{\"values\":[[1.0],[2.0]]}}";
    SeldonMessage.Builder builder = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
    SeldonMessage request = builder.build();

    String json2 = ProtoBufUtils.toJson(request);

    System.out.println(json2);
  }

  @Test
  public void defaultRequest() throws InvalidProtocolBufferException {
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

    System.out.println(json);

    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(b2, json);

    SeldonMessage request2 = b2.build();

    String json2 = ProtoBufUtils.toJson(request2);

    System.out.println(json2);

    Assert.assertEquals(json, json2);
  }

  @Test
  public void customBytesRequest() throws InvalidProtocolBufferException {
    String customData = "{\"c\":1.0}";
    SeldonMessage.Builder b = SeldonMessage.newBuilder();
    b.setBinData(ByteString.copyFrom(customData.getBytes()));
    SeldonMessage request = b.build();

    String json = ProtoBufUtils.toJson(request);

    System.out.println(json);

    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(b2, json);

    SeldonMessage request2 = b2.build();
    String custom = request2.getBinData().toString(StandardCharsets.UTF_8);
    System.out.println(custom);

    String json2 = ProtoBufUtils.toJson(request2);

    System.out.println(json2);

    Assert.assertEquals(json, json2);
  }

  @Test
  public void customStringRequest() throws InvalidProtocolBufferException {
    String customData = "{\"c\":1.0}";
    SeldonMessage.Builder b = SeldonMessage.newBuilder();
    b.setStrData(customData);
    SeldonMessage request = b.build();

    String json = ProtoBufUtils.toJson(request);

    System.out.println(json);

    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    ProtoBufUtils.updateMessageBuilderFromJson(b2, json);

    SeldonMessage request2 = b2.build();

    String json2 = ProtoBufUtils.toJson(request2);

    System.out.println(json2);

    Assert.assertEquals(json, json2);
  }
}
