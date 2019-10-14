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

import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitMethod;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class PredictorConfigBean {
  public final Map<PredictiveUnitType, List<PredictiveUnitMethod>> typeMethodsMap;
  public final Map<PredictiveUnitImplementation, PredictiveUnitImpl> nodeImplementationMap;

  @Autowired
  public PredictorConfigBean(
      SimpleModelUnit simpleModelUnit,
      SimpleRouterUnit simpleRouterUnit,
      AverageCombinerUnit averageCombinerUnit,
      RandomABTestUnit randomABTestUnit) {

    // ---------------------------
    // DEFINITION OF DEFAULT TYPES
    // ---------------------------
    typeMethodsMap = new HashMap<PredictiveUnitType, List<PredictiveUnitMethod>>();

    // MODEL -> TRANSFORM INPUT
    List<PredictiveUnitMethod> modelMethods = new ArrayList<PredictiveUnitMethod>();
    modelMethods.add(PredictiveUnitMethod.TRANSFORM_INPUT);
    modelMethods.add(PredictiveUnitMethod.SEND_FEEDBACK);
    typeMethodsMap.put(PredictiveUnitType.MODEL, modelMethods);

    // TRANSFORMER -> TRANSFORM INPUT
    List<PredictiveUnitMethod> transformerMethods = new ArrayList<PredictiveUnitMethod>();
    transformerMethods.add(PredictiveUnitMethod.TRANSFORM_INPUT);
    typeMethodsMap.put(PredictiveUnitType.TRANSFORMER, transformerMethods);

    // OUTPUT TRANSFORMER -> TRANSFORM OUTPUT
    List<PredictiveUnitMethod> outTransformerMethods = new ArrayList<PredictiveUnitMethod>();
    outTransformerMethods.add(PredictiveUnitMethod.TRANSFORM_OUTPUT);
    typeMethodsMap.put(PredictiveUnitType.OUTPUT_TRANSFORMER, outTransformerMethods);

    // ROUTER -> ROUTE, SEND FEEDBACK
    List<PredictiveUnitMethod> routerMethods = new ArrayList<PredictiveUnitMethod>();
    routerMethods.add(PredictiveUnitMethod.ROUTE);
    routerMethods.add(PredictiveUnitMethod.SEND_FEEDBACK);
    typeMethodsMap.put(PredictiveUnitType.ROUTER, routerMethods);

    // COMBINER -> AGGREGATE
    List<PredictiveUnitMethod> combinerMethods = new ArrayList<PredictiveUnitMethod>();
    combinerMethods.add(PredictiveUnitMethod.AGGREGATE);
    typeMethodsMap.put(PredictiveUnitType.COMBINER, combinerMethods);

    // -------------------------
    // HARDCODED IMPLEMENTATIONS
    // -------------------------
    nodeImplementationMap = new HashMap<PredictiveUnitImplementation, PredictiveUnitImpl>();
    nodeImplementationMap.put(PredictiveUnitImplementation.AVERAGE_COMBINER, averageCombinerUnit);
    nodeImplementationMap.put(PredictiveUnitImplementation.SIMPLE_MODEL, simpleModelUnit);
    nodeImplementationMap.put(PredictiveUnitImplementation.SIMPLE_ROUTER, simpleRouterUnit);
    nodeImplementationMap.put(PredictiveUnitImplementation.RANDOM_ABTEST, randomABTestUnit);
  }

  public boolean hasMethod(PredictiveUnitState state) {
    return (state.implementation == PredictiveUnitImplementation.UNKNOWN_IMPLEMENTATION
        || !nodeImplementationMap.containsKey(state.implementation));
  }

  public PredictiveUnitImpl getImplementation(PredictiveUnitState state) {
    if (state.implementation != PredictiveUnitImplementation.UNKNOWN_IMPLEMENTATION) {
      return nodeImplementationMap.get(state.implementation);
    }
    return null;
  }

  public Boolean hasMethod(PredictiveUnitMethod method, PredictiveUnitState state) {
    if (state.implementation != PredictiveUnitImplementation.UNKNOWN_IMPLEMENTATION
        && nodeImplementationMap.containsKey(state.implementation)) {
      return false;
    }
    if (state.type == PredictiveUnitType.UNKNOWN_TYPE) {
      return state.methods.contains(method);
    } else {
      return typeMethodsMap.get(state.type).contains(method);
    }
  }
}
