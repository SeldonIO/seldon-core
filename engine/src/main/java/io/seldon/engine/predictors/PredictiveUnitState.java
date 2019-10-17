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

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import io.kubernetes.client.proto.V1.Container;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.Parameter;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitMethod;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.apache.commons.lang3.StringUtils;
import org.springframework.beans.factory.annotation.Autowired;

@JsonIgnoreProperties({"children", "cluster_resources", "id", "subtype", "type"})
public class PredictiveUnitState {
  public String name;
  public Endpoint endpoint;
  public List<PredictiveUnitState> children = new ArrayList<>();
  public Map<String, PredictiveUnitParameterInterface> parameters;
  public String image = "";
  public String imageName;
  public String imageVersion;
  public PredictiveUnitType type;
  public PredictiveUnitImplementation implementation;
  public List<PredictiveUnitMethod> methods;

  @Autowired public PredictorBean predictorBean;

  public PredictiveUnitState() {}

  public PredictiveUnitState(
      String name,
      Endpoint endpoint,
      List<PredictiveUnitState> children,
      Map<String, PredictiveUnitParameterInterface> parameters,
      String imageName,
      String imageVersion,
      PredictiveUnitType type,
      PredictiveUnitImplementation implementation) {
    this.name = name;
    this.endpoint = endpoint;
    this.children = children;
    this.parameters = parameters;
    this.imageName = imageName;
    this.imageVersion = imageVersion;
    if (!StringUtils.isEmpty(imageName) && !StringUtils.isEmpty(imageVersion))
      this.image = imageName + ":" + imageVersion;
    else if (!StringUtils.isEmpty(imageName)) this.image = imageName;
    else this.image = "";
    this.type = type;
    this.implementation = implementation;
  }

  public PredictiveUnitState(PredictiveUnit predictiveUnit, Map<String, Container> containersMap) {
    this.name = predictiveUnit.getName();
    this.endpoint = predictiveUnit.getEndpoint();
    this.parameters = deserializeParameters(predictiveUnit.getParametersList());

    if (containersMap.containsKey(name)) {
      this.image = containersMap.get(name).getImage();
      int i = image.lastIndexOf(":");
      if (i >= 0) {
        this.imageName = StringUtils.substring(image, 0, i);
        this.imageVersion = StringUtils.substring(image, i + 1);
      } else {
        this.imageName = image;
        this.imageVersion = "";
      }
    }

    this.children = new ArrayList<PredictiveUnitState>();

    for (PredictiveUnit childUnit : predictiveUnit.getChildrenList()) {
      this.children.add(new PredictiveUnitState(childUnit, containersMap));
    }

    this.type = predictiveUnit.getType();
    this.implementation = predictiveUnit.getImplementation();
    this.methods = predictiveUnit.getMethodsList();
  }

  public static Map<String, PredictiveUnitParameterInterface> deserializeParameters(
      List<Parameter> Parameters) {
    Map<String, PredictiveUnitParameterInterface> paramsMap = new HashMap<>();
    for (Parameter Parameter : Parameters) {
      paramsMap.put(Parameter.getName(), PredictiveUnitParameter.fromParameter(Parameter));
    }
    return paramsMap;
  }

  public void addChild(PredictiveUnitState predictiveUnitState) {
    this.children.add(predictiveUnitState);
  }
}
