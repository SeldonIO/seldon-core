package io.seldon.engine.predictors;

import java.util.List;
import java.util.ArrayList;
import java.lang.reflect.Method;
import java.lang.reflect.InvocationTargetException;
import java.lang.NoSuchMethodException;
import java.lang.IllegalAccessException;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;

import org.springframework.beans.factory.annotation.Autowired;

import io.kubernetes.client.proto.V1.Container;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.engine.pb.ProtoBufUtils;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Value;

public class PredictiveUnitBeanTest {
	
	@Test
	public void testMergeMetaList() throws InvalidProtocolBufferException, NoSuchMethodException, IllegalAccessException, InvocationTargetException
	{
        PredictiveUnitImpl pu = new PredictiveUnitBean();
        Method mergeMeta = PredictiveUnitBean.class.getDeclaredMethod("mergeMeta", SeldonMessage.class, List.class);
        mergeMeta.setAccessible(true);

		DefaultData.Builder d1 = DefaultData.newBuilder();
		SeldonMessage.Builder b1 = SeldonMessage.newBuilder();
		Value v1 = Value.newBuilder().setStringValue("one").build();
		Meta m1 = Meta.newBuilder().putTags("key", v1).build();
		b1.setData(d1.build()).setMeta(m1);
        SeldonMessage s1 = b1.build();

		DefaultData.Builder d2 = DefaultData.newBuilder();
		SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
		Value v2 = Value.newBuilder().setStringValue("two").build();
		Meta m2 = Meta.newBuilder().putTags("key", v2).build();
		b2.setData(d2.build()).setMeta(m2);
        SeldonMessage s2 = b2.build();

        List<SeldonMessage> messages = new ArrayList<SeldonMessage>();
        messages.add(s2);

        SeldonMessage r1 = (SeldonMessage) mergeMeta.invoke(pu, s1, messages);

		String r1str = ProtoBufUtils.toJson(r1.getMeta());
		String j2 = ProtoBufUtils.toJson(s2.getMeta());
        System.out.println(r1str);
        System.out.println(j2);
		Assert.assertEquals(r1str, j2);


		DefaultData.Builder d3 = DefaultData.newBuilder();
		SeldonMessage.Builder b3 = SeldonMessage.newBuilder();
		Value v3 = Value.newBuilder().setStringValue("three").build();
		Meta m3 = Meta.newBuilder().putTags("key", v3).build();
		b3.setData(d3.build()).setMeta(m3);
        SeldonMessage s3 = b3.build();

        messages.add(s3);

        SeldonMessage r2 = (SeldonMessage) mergeMeta.invoke(pu, s1, messages);

		String r2str = ProtoBufUtils.toJson(r2.getMeta());
		String j3 = ProtoBufUtils.toJson(s3.getMeta());
        System.out.println(r2str);
        System.out.println(j3);
		Assert.assertEquals(r2str, j3);
	}

	@Test
	public void testMergeMeta() throws InvalidProtocolBufferException, NoSuchMethodException, IllegalAccessException, InvocationTargetException
	{
        PredictiveUnitImpl pu = new PredictiveUnitBean();
        Method mergeMeta = PredictiveUnitBean.class.getDeclaredMethod("mergeMeta", SeldonMessage.class, Meta.class);
        mergeMeta.setAccessible(true);

		DefaultData.Builder d1 = DefaultData.newBuilder();
		SeldonMessage.Builder b1 = SeldonMessage.newBuilder();
		Value v1 = Value.newBuilder().setStringValue("one").build();
		Meta m1 = Meta.newBuilder().putTags("key", v1).build();
		b1.setData(d1.build()).setMeta(m1);
        SeldonMessage s1 = b1.build();

		DefaultData.Builder d2 = DefaultData.newBuilder();
		SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
		Value v2 = Value.newBuilder().setStringValue("two").build();
		Meta m2 = Meta.newBuilder().putTags("key", v2).build();
		b2.setData(d2.build()).setMeta(m2);
        SeldonMessage s2 = b2.build();

        SeldonMessage r1 = (SeldonMessage) mergeMeta.invoke(pu, s1, s2.getMeta());

		String r1str = ProtoBufUtils.toJson(r1.getMeta());
		String j2 = ProtoBufUtils.toJson(s2.getMeta());
        System.out.println(r1str);
        System.out.println(j2);
		Assert.assertEquals(r1str, j2);
	}
}


