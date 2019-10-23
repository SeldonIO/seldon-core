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
package io.seldon.engine.predictors;

import com.google.protobuf.InvalidProtocolBufferException;
import io.seldon.engine.exception.APIException;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import org.junit.Assert;
import org.junit.Test;
import org.ojalgo.matrix.PrimitiveMatrix;

public class RandomABTestUnitInternalTest {

  private int getBranchIndex(SeldonMessage routerReturn) {
    PrimitiveMatrix dataArray = PredictorUtils.getOJMatrix(routerReturn.getData());
    return dataArray.get(0).intValue();
  }

  @Test
  public void simpleCase()
      throws NoSuchMethodException, SecurityException, IllegalAccessException,
          IllegalArgumentException, InvalidProtocolBufferException {

    SeldonMessage request = SeldonMessage.newBuilder().build();

    PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    Map<String, PredictiveUnitParameterInterface> params = new HashMap<>();
    params.put("ratioA", ratioParam);

    PredictiveUnitState state =
        new PredictiveUnitState(
            "Cool_name",
            null,
            new ArrayList<PredictiveUnitState>(),
            params,
            null,
            null,
            null,
            PredictiveUnitImplementation.RANDOM_ABTEST);

    PredictiveUnitState childA =
        new PredictiveUnitState("A", null, null, null, null, null, null, null);
    PredictiveUnitState childB =
        new PredictiveUnitState("B", null, null, null, null, null, null, null);

    state.addChild(childA);
    state.addChild(childB);

    PredictiveUnitImpl predictiveUnit = new RandomABTestUnit();

    // The following values are from random seed 1337
    SeldonMessage msg1 = predictiveUnit.route(request, state);
    int routing1 = getBranchIndex(msg1);

    Assert.assertEquals(1, routing1);

    SeldonMessage msg2 = predictiveUnit.route(request, state);
    int routing2 = (int) getBranchIndex(msg2);

    Assert.assertEquals(0, routing2);

    SeldonMessage msg3 = predictiveUnit.route(request, state);
    int routing3 = (int) getBranchIndex(msg3);

    Assert.assertEquals(1, routing3);
  }

  @Test(expected = APIException.class)
  public void failureOneChild() throws Throwable {

    SeldonMessage request = SeldonMessage.newBuilder().build();

    PredictiveUnitParameter<Float> ratioParam = new PredictiveUnitParameter<Float>(0.5F);
    Map<String, PredictiveUnitParameterInterface> params = new HashMap<>();
    params.put("ratioA", ratioParam);

    PredictiveUnitState state =
        new PredictiveUnitState(
            "Cool_name",
            null,
            new ArrayList<PredictiveUnitState>(),
            params,
            null,
            null,
            null,
            PredictiveUnitImplementation.RANDOM_ABTEST);

    PredictiveUnitState childA =
        new PredictiveUnitState("A", null, null, null, null, null, null, null);

    state.addChild(childA);

    PredictiveUnitImpl predictiveUnit = new RandomABTestUnit();

    predictiveUnit.route(request, state);
  }
}
