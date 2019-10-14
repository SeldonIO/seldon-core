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

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import java.util.Random;
import org.springframework.stereotype.Component;

@Component
public class RandomABTestUnit extends PredictiveUnitImpl {

  Random rand = new Random(1337);

  public RandomABTestUnit() {}

  @Override
  public SeldonMessage route(SeldonMessage input, PredictiveUnitState state) {
    @SuppressWarnings("unchecked")
    PredictiveUnitParameter<Float> parameter =
        (PredictiveUnitParameter<Float>) state.parameters.get("ratioA");

    if (parameter == null) {
      throw new APIException(
          APIException.ApiExceptionType.ENGINE_INVALID_ABTEST, "Parameter 'ratioA' is missing.");
    }

    Float ratioA = parameter.value;
    Float comparator = rand.nextFloat();

    if (state.children.size() != 2) {
      throw new APIException(
          APIException.ApiExceptionType.ENGINE_INVALID_ABTEST,
          String.format("AB test has %d children ", state.children.size()));
    }

    // FIXME Possible bug : keySet is not ordered as per the definition of the AB test
    if (comparator <= ratioA) {
      // We select model A
      return SeldonMessage.newBuilder()
          .setData(
              DefaultData.newBuilder()
                  .setTensor(Tensor.newBuilder().addValues(0).addShape(1).addShape(1)))
          .build();
    } else {
      return SeldonMessage.newBuilder()
          .setData(
              DefaultData.newBuilder()
                  .setTensor(Tensor.newBuilder().addValues(1).addShape(1).addShape(1)))
          .build();
    }
  }
}
