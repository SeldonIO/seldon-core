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

// import static org.ojalgo.function.PrimitiveFunction.*;
import static org.ojalgo.function.constant.PrimitiveMath.DIVIDE;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultData;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import java.util.Iterator;
import java.util.List;
import org.ojalgo.matrix.PrimitiveMatrix;
import org.springframework.stereotype.Component;

@Component
public class AverageCombinerUnit extends PredictiveUnitImpl {

  public AverageCombinerUnit() {}

  @Override
  public SeldonMessage aggregate(List<SeldonMessage> outputs, PredictiveUnitState state) {

    if (outputs.size() == 0) {
      throw new APIException(
          APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,
          String.format("Combiner received no inputs"));
    }

    int[] shape = PredictorUtils.getShape(outputs.get(0).getData());

    if (shape == null) {
      throw new APIException(
          APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,
          String.format("Combiner cannot extract data shape"));
    }

    if (shape.length != 2) {
      throw new APIException(
          APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,
          String.format("Combiner received data that is not 2 dimensional"));
    }
    PrimitiveMatrix.DenseReceiver currentSum =
        PrimitiveMatrix.FACTORY.makeDense(shape[0], shape[1]);
    SeldonMessage.Builder respBuilder = SeldonMessage.newBuilder();

    for (Iterator<SeldonMessage> i = outputs.iterator(); i.hasNext(); ) {
      DefaultData inputData = i.next().getData();
      int[] inputShape = PredictorUtils.getShape(inputData);
      if (inputShape == null) {
        throw new APIException(
            APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,
            String.format("Combiner cannot extract data shape"));
      }
      if (inputShape.length != 2) {
        throw new APIException(
            APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,
            String.format("Combiner received data that is not 2 dimensional"));
      }
      if (inputShape[0] != shape[0]) {
        throw new APIException(
            APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,
            String.format("Expected batch length %d but found %d", shape[0], inputShape[0]));
      }
      if (inputShape[1] != shape[1]) {
        throw new APIException(
            APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,
            String.format("Expected batch length %d but found %d", shape[1], inputShape[1]));
      }
      PredictorUtils.add(inputData, currentSum);
    }
    currentSum.modifyAll(DIVIDE.by(outputs.size()));

    DefaultData newData = PredictorUtils.updateData(outputs.get(0).getData(), currentSum.get());
    respBuilder.setData(newData);
    respBuilder.setMeta(outputs.get(0).getMeta());
    respBuilder.setStatus(outputs.get(0).getStatus());

    return respBuilder.build();
  }
}
