package io.seldon.engine.pb;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

import org.junit.Assert;
import org.junit.Test;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Value;

import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.MetaDef;
import io.seldon.protos.PredictionProtos.Tensor;

public class TestPredictionProto {

	@Test
	public void parse_json_extra_fields() throws InvalidProtocolBufferException
	{
		String json = "{\"x\":1.0,\"request\":{\"values\":[[1.0],[2.0]]}}";
		RequestDef.Builder builder = RequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		RequestDef request = builder.build();
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}
	
	
	@Test
	public void parse_custom_json() throws InvalidProtocolBufferException
	{
		String json = "{\"request\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
		RequestDef.Builder builder = RequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		RequestDef request = builder.build();
		
		Assert.assertEquals(2, request.getData().getNdarray().getValuesCount());
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}
	
	@Test
	public void parse_tags_array() throws InvalidProtocolBufferException
	{
		String json = "{\"meta\":{\"tags\":{\"user\":[\"a\",\"b\"],\"gender\":\"female\"}},\"request\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
		RequestDef.Builder builder = RequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		RequestDef request = builder.build();
		
		Assert.assertEquals(2, request.getData().getNdarray().getValuesCount());
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}

	
	
	@Test
	public void parse_json() throws InvalidProtocolBufferException
	{
		String json = "{\"request\":{\"values\":[[1.0],[2.0]]}}";
		RequestDef.Builder builder = RequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		RequestDef request = builder.build();
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}
	
	
	@Test
	public void defaultRequest() throws InvalidProtocolBufferException
	{
		String[] features = {"a","b"};
		Double[] values = {1.0,2.0,1.5,2.2};
		DefaultDataDef.Builder defB = DefaultDataDef.newBuilder();
		defB.addAllNames( Arrays.asList(features) );
		defB.setTensor(Tensor.newBuilder().addShape(1).addShape(values.length).addAllValues(Arrays.asList(values)).build());
		RequestDef.Builder b = RequestDef.newBuilder();
		Value v;
		Value v1 = Value.newBuilder().setNumberValue(1.0).build();
		
		b.setData(defB.build()).setMeta(MetaDef.newBuilder().putTags("key", v1).build());
		RequestDef request = b.build();
		
		String json = ProtoBufUtils.toJson(request);
		
		System.out.println(json);
		
		RequestDef.Builder b2 = RequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
		
		RequestDef request2 = b2.build();
		
		String json2 = ProtoBufUtils.toJson(request2);
		
		System.out.println(json2);
		
		Assert.assertEquals(json, json2);
		
	}
	
	@Test
	public void customBytesRequest() throws InvalidProtocolBufferException
	{
		String customData = "{\"c\":1.0}";
		RequestDef.Builder b = RequestDef.newBuilder();
		b.setBinData(ByteString.copyFrom(customData.getBytes()));
		RequestDef request = b.build();
		
		String json = ProtoBufUtils.toJson(request);
		
		System.out.println(json);
		
		RequestDef.Builder b2 = RequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
		
		RequestDef request2 = b2.build();
		String custom = request2.getBinData().toString(StandardCharsets.UTF_8);
		System.out.println(custom);
		
		String json2 = ProtoBufUtils.toJson(request2);
		
		System.out.println(json2);
		
		Assert.assertEquals(json, json2);
	}
	
	@Test 
	public void customStringRequest() throws InvalidProtocolBufferException
	{
		String customData = "{\"c\":1.0}";
		RequestDef.Builder b = RequestDef.newBuilder();
		b.setStrData(customData);
		RequestDef request = b.build();
		
		String json = ProtoBufUtils.toJson(request);
		
		System.out.println(json);
		
		RequestDef.Builder b2 = RequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
		
		RequestDef request2 = b2.build();
		
		String json2 = ProtoBufUtils.toJson(request2);
		
		System.out.println(json2);
		
		Assert.assertEquals(json, json2);
	}
	
}
