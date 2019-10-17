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

import io.seldon.protos.DeploymentProtos.Parameter;

public class PredictiveUnitParameter<T> extends PredictiveUnitParameterInterface {
  public T value;

  public PredictiveUnitParameter(T value) {
    this.value = value;
  }

  public static PredictiveUnitParameterInterface fromParameter(Parameter Parameter) {
    switch (Parameter.getType()) {
      case DOUBLE:
        Double valueDouble = Double.parseDouble(Parameter.getValue());
        return new PredictiveUnitParameter<Double>(valueDouble);
      case FLOAT:
        Float valueFloat = Float.parseFloat(Parameter.getValue());
        return new PredictiveUnitParameter<Float>(valueFloat);
      case INT:
        Integer valueInt = Integer.parseInt(Parameter.getValue());
        return new PredictiveUnitParameter<Integer>(valueInt);
      case STRING:
        return new PredictiveUnitParameter<String>(Parameter.getValue());
      default:
        return new PredictiveUnitParameter<String>(Parameter.getValue());
    }
  }
}
