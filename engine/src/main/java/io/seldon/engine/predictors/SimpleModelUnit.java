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
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.Metric;
import io.seldon.protos.PredictionProtos.Metric.MetricType;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Status;
import io.seldon.protos.PredictionProtos.Tensor;
import java.util.Arrays;
import org.springframework.stereotype.Component;

@Component
public class SimpleModelUnit extends PredictiveUnitImpl {

  public SimpleModelUnit() {}

  public static final Double[] values = {0.1, 0.9, 0.5};
  public static final String[] classes = {"class0", "class1", "class2"};

  @Override
  public SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state) {
    SeldonMessage.Builder builder =
        SeldonMessage.newBuilder()
            .setStatus(Status.newBuilder().setStatus(Status.StatusFlag.SUCCESS).build())
            .setMeta(
                Meta.newBuilder()
                    .addMetrics(
                        Metric.newBuilder()
                            .setKey("mymetric_counter")
                            .setType(MetricType.COUNTER)
                            .setValue(1))
                    .addMetrics(
                        Metric.newBuilder()
                            .setKey("mymetric_gauge")
                            .setType(MetricType.GAUGE)
                            .setValue(100))
                    .addMetrics(
                        Metric.newBuilder()
                            .setKey("mymetric_timer")
                            .setType(MetricType.TIMER)
                            .setValue(22.1F)));

    // echo in case of strData and binData
    if (input.getDataOneofCase().equals(SeldonMessage.DataOneofCase.BINDATA)) {
      builder.setBinData(input.getBinData());
    } else if (input.getDataOneofCase().equals(SeldonMessage.DataOneofCase.STRDATA)) {
      builder.setStrData(input.getStrData());
    } else {
      builder.setData(
          DefaultData.newBuilder()
              .addAllNames(Arrays.asList(classes))
              .setTensor(
                  Tensor.newBuilder()
                      .addShape(1)
                      .addShape(values.length)
                      .addAllValues(Arrays.asList(values))));
    }

    SeldonMessage output = builder.build();
    System.out.println("Model " + state.name + " finishing computations");
    return output;
  }
}
