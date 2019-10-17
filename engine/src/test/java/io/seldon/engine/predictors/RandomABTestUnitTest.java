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
import io.seldon.protos.DeploymentProtos.Parameter;
import io.seldon.protos.DeploymentProtos.Parameter.ParameterType;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.DeploymentProtos.SeldonPodSpec;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import java.util.concurrent.ExecutionException;
import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
@SpringBootTest
public class RandomABTestUnitTest {
  @Autowired PredictorBean predictorBean;

  @Test
  public void simpleTest()
      throws InterruptedException, ExecutionException, InvalidProtocolBufferException {
    PredictorSpec.Builder predictorSpecBuilder = PredictorSpec.newBuilder();

    predictorSpecBuilder.setName("p1");
    predictorSpecBuilder.setReplicas(1);
    predictorSpecBuilder.addComponentSpecs(SeldonPodSpec.newBuilder());

    PredictiveUnit.Builder predictiveUnitBuilder = PredictiveUnit.newBuilder();
    predictiveUnitBuilder.setName("1");
    predictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
    predictiveUnitBuilder.setImplementation(
        PredictiveUnit.PredictiveUnitImplementation.SIMPLE_MODEL);
    PredictiveUnit pu1 = predictiveUnitBuilder.build();

    predictiveUnitBuilder = PredictiveUnit.newBuilder();
    predictiveUnitBuilder.setName("2");
    predictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.MODEL);
    predictiveUnitBuilder.setImplementation(
        PredictiveUnit.PredictiveUnitImplementation.SIMPLE_MODEL);
    PredictiveUnit pu2 = predictiveUnitBuilder.build();

    predictiveUnitBuilder = PredictiveUnit.newBuilder();
    predictiveUnitBuilder.setName("3");
    predictiveUnitBuilder.setType(PredictiveUnit.PredictiveUnitType.ROUTER);
    predictiveUnitBuilder.setImplementation(
        PredictiveUnit.PredictiveUnitImplementation.RANDOM_ABTEST);
    predictiveUnitBuilder.addChildren(pu1);
    predictiveUnitBuilder.addChildren(pu2);
    Parameter.Builder pBuilder =
        Parameter.newBuilder().setName("ratioA").setValue("0.5").setType(ParameterType.FLOAT);
    predictiveUnitBuilder.addParameters(pBuilder.build());
    PredictiveUnit pu3 = predictiveUnitBuilder.build();

    predictorSpecBuilder.setGraph(pu3);

    PredictorSpec predictor = predictorSpecBuilder.build();

    PredictorState predictorState = predictorBean.predictorStateFromPredictorSpec(predictor);

    SeldonMessage p = SeldonMessage.newBuilder().build();

    SeldonMessage predictorReturn = predictorBean.predict(p, predictorState);

    Assert.assertEquals(
        (double) SimpleModelUnit.values[0], predictorReturn.getData().getTensor().getValues(0), 0);
    Assert.assertEquals(
        (double) SimpleModelUnit.values[1], predictorReturn.getData().getTensor().getValues(1), 0);
    Assert.assertEquals(
        (double) SimpleModelUnit.values[2], predictorReturn.getData().getTensor().getValues(2), 0);
  }
}
