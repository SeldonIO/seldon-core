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
		return predictiveUnitBean.getOutput(request, rootState);
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