package io.seldon.engine.predictors;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Value;
import io.seldon.engine.pb.ProtoBufUtils;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.List;
import org.junit.Assert;
import org.junit.Test;

public class PredictiveUnitBeanTest {

  @Test
  public void testMergeMetaList()
      throws InvalidProtocolBufferException, NoSuchMethodException, IllegalAccessException,
          InvocationTargetException {
    PredictiveUnitImpl pu = new PredictiveUnitBean();
    Method mergeMeta =
        PredictiveUnitBean.class.getDeclaredMethod(
            "mergeMeta", SeldonMessage.class, List.class, String.class);
    mergeMeta.setAccessible(true);

    final String puid = "puid";

    DefaultData.Builder d1 = DefaultData.newBuilder();
    SeldonMessage.Builder b1 = SeldonMessage.newBuilder();
    Value v1 = Value.newBuilder().setStringValue("one").build();
    Meta m1 = Meta.newBuilder().putTags("key", v1).setPuid(puid).build();
    b1.setData(d1.build()).setMeta(m1);
    SeldonMessage s1 = b1.build();

    DefaultData.Builder d2 = DefaultData.newBuilder();
    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    Value v2 = Value.newBuilder().setStringValue("two").build();
    Meta m2 = Meta.newBuilder().putTags("key", v2).setPuid(puid).build();
    b2.setData(d2.build()).setMeta(m2);
    SeldonMessage s2 = b2.build();

    List<SeldonMessage> messages = new ArrayList<SeldonMessage>();
    messages.add(s2);

    SeldonMessage r1 = (SeldonMessage) mergeMeta.invoke(pu, s1, messages, puid);

    String r1str = ProtoBufUtils.toJson(r1.getMeta());
    String j1 = ProtoBufUtils.toJson(s1.getMeta());
    Assert.assertEquals(r1str, j1);

    DefaultData.Builder d3 = DefaultData.newBuilder();
    SeldonMessage.Builder b3 = SeldonMessage.newBuilder();
    Value v3 = Value.newBuilder().setStringValue("three").build();
    Meta m3 = Meta.newBuilder().putTags("key", v3).setPuid(puid).build();
    b3.setData(d3.build()).setMeta(m3);
    SeldonMessage s3 = b3.build();

    messages.add(s3);

    SeldonMessage r2 = (SeldonMessage) mergeMeta.invoke(pu, s1, messages, puid);

    String r2str = ProtoBufUtils.toJson(r2.getMeta());
    Assert.assertEquals(r2str, j1);
  }

  @Test
  public void testMergeMeta()
      throws InvalidProtocolBufferException, NoSuchMethodException, IllegalAccessException,
          InvocationTargetException {
    PredictiveUnitImpl pu = new PredictiveUnitBean();
    Method mergeMeta =
        PredictiveUnitBean.class.getDeclaredMethod(
            "mergeMeta", SeldonMessage.class, SeldonMessage.class, String.class);
    mergeMeta.setAccessible(true);

    final String puid = "id";

    DefaultData.Builder d1 = DefaultData.newBuilder();
    SeldonMessage.Builder b1 = SeldonMessage.newBuilder();
    Value v1 = Value.newBuilder().setStringValue("one").build();
    Meta m1 = Meta.newBuilder().putTags("key", v1).setPuid(puid).build();
    b1.setData(d1.build()).setMeta(m1);
    SeldonMessage s1 = b1.build();

    DefaultData.Builder d2 = DefaultData.newBuilder();
    SeldonMessage.Builder b2 = SeldonMessage.newBuilder();
    Value v2 = Value.newBuilder().setStringValue("two").build();
    Meta m2 = Meta.newBuilder().putTags("key", v2).setPuid(puid).build();
    b2.setData(d2.build()).setMeta(m2);
    SeldonMessage s2 = b2.build();

    SeldonMessage r1 = (SeldonMessage) mergeMeta.invoke(pu, s1, s2, puid);

    String r1str = ProtoBufUtils.toJson(r1.getMeta());
    String j1 = ProtoBufUtils.toJson(s1.getMeta());
    Assert.assertEquals(r1str, j1);
  }
}
