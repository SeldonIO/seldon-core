package io.seldon.engine.pb;

import io.seldon.engine.predictors.PredictorUtils;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Status;
import io.seldon.protos.PredictionProtos.Tensor;
import java.util.Arrays;
import org.junit.Assert;
import org.junit.Test;
import org.ojalgo.matrix.PrimitiveMatrix;

public class TestMatrixOps {

  @Test
  public void testOj() {
    String[] names = {"c", "d"};
    Double[] values1 = {1.0, 2.0, 3.0, 4.0};
    SeldonMessage m =
        SeldonMessage.newBuilder()
            .setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
            .setData(
                DefaultData.newBuilder()
                    .addAllNames(Arrays.asList(names))
                    .setTensor(
                        Tensor.newBuilder()
                            .addShape(2)
                            .addShape(2)
                            .addAllValues(Arrays.asList(values1))
                            .build())
                    .build())
            .build();
    PrimitiveMatrix p = PredictorUtils.getOJMatrix(m.getData());
    Assert.assertEquals(values1[0], p.get(0, 0));
    Assert.assertEquals(values1[1], p.get(0, 1));
    Assert.assertEquals(values1[2], p.get(1, 0));
    Assert.assertEquals(values1[3], p.get(1, 1));

    Assert.assertEquals(values1[0], p.get(0));
  }
}
