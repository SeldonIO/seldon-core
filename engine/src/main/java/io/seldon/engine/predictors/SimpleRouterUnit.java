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

import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Tensor;
import org.springframework.stereotype.Component;

@Component
public class SimpleRouterUnit extends PredictiveUnitImpl {

  public SimpleRouterUnit() {}

  @Override
  public SeldonMessage route(SeldonMessage input, PredictiveUnitState state) {
    return SeldonMessage.newBuilder()
        .setData(
            DefaultData.newBuilder()
                .setTensor(Tensor.newBuilder().addValues(0).addShape(1).addShape(1)))
        .build();
  }
}
