/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitMethod;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

import io.kubernetes.client.proto.V1.Container;


@Component
public class PredictorBean {
	
	@Autowired
	PredictiveUnitBean predictiveUnitBean;
	
    @Autowired
	public PredictorBean(){
    }
   
	public SeldonMessage predict(SeldonMessage request, PredictorState predictorState) throws InterruptedException, ExecutionException, InvalidProtocolBufferException

	{
		PredictiveUnitState rootState = predictorState.rootState;
		return this.predictiveUnitBean.getOutput(request, rootState);
	}
	
	public void sendFeedback(Feedback feedback, PredictorState predictorState) throws InterruptedException, ExecutionException, InvalidProtocolBufferException

	{
		PredictiveUnitState rootState = predictorState.rootState;
		predictiveUnitBean.sendFeedback(feedback, rootState);
		return;
	}
	
	//TODO
	public PredictorState predictorStateFromPredictorSpec(PredictorSpec predictorSpec){

        // Boolean enabled = PredictorSpec.getEnabled();
		Boolean enabled = true;
		PredictiveUnit rootUnit = predictorSpec.getGraph();
		Map<String,Container> containersMap = new HashMap<String,Container>();
		
		for (Container container : predictorSpec.getComponentSpec().getSpec().getContainersList()){
			containersMap.put(container.getName(), container);
		}
		
		PredictiveUnitState rootState = new PredictiveUnitState(rootUnit,containersMap);
		
		return new PredictorState(rootUnit.getName(),rootState, enabled);
	
	}
	
	
}