package io.seldon.engine.pb;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

import org.junit.Assert;
import org.junit.Test;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Value;

import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionRequestMetaDef;
import io.seldon.protos.PredictionProtos.Tensor;

public class TestPredictionProto {

	@Test
	public void parse_json_extra_fields() throws InvalidProtocolBufferException
	{
		String json = "{\"x\":1.0,\"request\":{\"values\":[[1.0],[2.0]]}}";
		PredictionRequestDef.Builder builder = PredictionRequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		PredictionRequestDef request = builder.build();
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}
	
	
	@Test
	public void parse_custom_json() throws InvalidProtocolBufferException
	{
		String json = "{\"request\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
		PredictionRequestDef.Builder builder = PredictionRequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		PredictionRequestDef request = builder.build();
		
		Assert.assertEquals(2, request.getRequest().getNdarray().getValuesCount());
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}
	
	@Test
	public void parse_tags_array() throws InvalidProtocolBufferException
	{
		String json = "{\"meta\":{\"tags\":{\"user\":[\"a\",\"b\"],\"gender\":\"female\"}},\"request\":{\"ndarray\":[[1.0,2.0],[3.0,4.0]]}}";
		PredictionRequestDef.Builder builder = PredictionRequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		PredictionRequestDef request = builder.build();
		
		Assert.assertEquals(2, request.getRequest().getNdarray().getValuesCount());
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}

	
	
	@Test
	public void parse_json() throws InvalidProtocolBufferException
	{
		String json = "{\"request\":{\"values\":[[1.0],[2.0]]}}";
		PredictionRequestDef.Builder builder = PredictionRequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(builder, json);
		PredictionRequestDef request = builder.build();
		
		String json2 = ProtoBufUtils.toJson(request);
		
		System.out.println(json2);
	}
	
	
	@Test
	public void defaultRequest() throws InvalidProtocolBufferException
	{
		String[] features = {"a","b"};
		Double[] values = {1.0,2.0,1.5,2.2};
		DefaultDataDef.Builder defB = DefaultDataDef.newBuilder();
		defB.addAllFeatures( Arrays.asList(features) );
		defB.setTensor(Tensor.newBuilder().addShape(1).addShape(values.length).addAllValues(Arrays.asList(values)).build());
		PredictionRequestDef.Builder b = PredictionRequestDef.newBuilder();
		Value v;
		Value v1 = Value.newBuilder().setNumberValue(1.0).build();
		
		b.setRequest(defB.build()).setMeta(PredictionRequestMetaDef.newBuilder().putTags("key", v1).build());
		PredictionRequestDef request = b.build();
		
		String json = ProtoBufUtils.toJson(request);
		
		System.out.println(json);
		
		PredictionRequestDef.Builder b2 = PredictionRequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
		
		PredictionRequestDef request2 = b2.build();
		
		String json2 = ProtoBufUtils.toJson(request2);
		
		System.out.println(json2);
		
		Assert.assertEquals(json, json2);
		
	}
	
	@Test
	public void customBytesRequest() throws InvalidProtocolBufferException
	{
		String customData = "{\"c\":1.0}";
		PredictionRequestDef.Builder b = PredictionRequestDef.newBuilder();
		b.setBinRequest(ByteString.copyFrom(customData.getBytes()));
		PredictionRequestDef request = b.build();
		
		String json = ProtoBufUtils.toJson(request);
		
		System.out.println(json);
		
		PredictionRequestDef.Builder b2 = PredictionRequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
		
		PredictionRequestDef request2 = b2.build();
		String custom = request2.getBinRequest().toString(StandardCharsets.UTF_8);
		System.out.println(custom);
		
		String json2 = ProtoBufUtils.toJson(request2);
		
		System.out.println(json2);
		
		Assert.assertEquals(json, json2);
	}
	
	@Test 
	public void customStringRequest() throws InvalidProtocolBufferException
	{
		String customData = "{\"c\":1.0}";
		PredictionRequestDef.Builder b = PredictionRequestDef.newBuilder();
		b.setStrRequest(customData);
		PredictionRequestDef request = b.build();
		
		String json = ProtoBufUtils.toJson(request);
		
		System.out.println(json);
		
		PredictionRequestDef.Builder b2 = PredictionRequestDef.newBuilder();
		ProtoBufUtils.updateMessageBuilderFromJson(b2, json);
		
		PredictionRequestDef request2 = b2.build();
		
		String json2 = ProtoBufUtils.toJson(request2);
		
		System.out.println(json2);
		
		Assert.assertEquals(json, json2);
	}
	
}
