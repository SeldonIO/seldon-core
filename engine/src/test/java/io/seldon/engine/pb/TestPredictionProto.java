package io.seldon.engine.pb;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

import org.junit.Assert;
import org.junit.Test;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionRequestMetaDef;

public class TestPredictionProto {

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
		defB.addAllKeys( Arrays.asList(features) );
		defB.addAllValues(Arrays.asList(values));
		PredictionRequestDef.Builder b = PredictionRequestDef.newBuilder();
		b.setRequest(defB.build()).setMeta(PredictionRequestMetaDef.newBuilder().putTags("key1", "val1").build());
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
