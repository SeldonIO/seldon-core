package io.seldon.engine.predictors;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ExecutionException;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

import io.kubernetes.client.proto.V1.Container;


@Component
public class PredictorBean {

    public final Map<PredictiveUnitType,PredictiveUnitBean> nodeTypeMap;
    public final Map<PredictiveUnitImplementation,PredictiveUnitBean> nodeImplementationMap;
	
    @Autowired
	public PredictorBean(
			ModelUnit modelUnit, 
			RouterUnit routerUnit, 
			CombinerUnit combinerUnit, 
			SimpleModelUnit simpleModelUnit, 
			SimpleRouterUnit simpleRouterUnit,
			AverageCombinerUnit averageCombinerUnit,
			TransformerUnit transformerUnit,
			RandomABTestUnit randomABTestUnit) {
        nodeTypeMap = new HashMap<PredictiveUnitType,PredictiveUnitBean>();
        nodeTypeMap.put(PredictiveUnitType.MODEL, modelUnit);
        nodeTypeMap.put(PredictiveUnitType.TRANSFORMER, transformerUnit);
        nodeTypeMap.put(PredictiveUnitType.ROUTER, routerUnit);
        nodeTypeMap.put(PredictiveUnitType.COMBINER, combinerUnit);
        
        nodeImplementationMap = new HashMap<PredictiveUnitImplementation,PredictiveUnitBean>();
        nodeImplementationMap.put(PredictiveUnitImplementation.AVERAGE_COMBINER, averageCombinerUnit);
        nodeImplementationMap.put(PredictiveUnitImplementation.SIMPLE_MODEL, simpleModelUnit);
        nodeImplementationMap.put(PredictiveUnitImplementation.SIMPLE_ROUTER, simpleRouterUnit);
        nodeImplementationMap.put(PredictiveUnitImplementation.RANDOM_ABTEST, randomABTestUnit);
    }
   
	public SeldonMessage predict(SeldonMessage request, PredictorState predictorState) throws InterruptedException, ExecutionException

	{
		PredictiveUnitState rootState = predictorState.rootState;
		return rootState.predictiveUnitBean.getOutput(request, rootState);
	}
	
	public void sendFeedback(Feedback feedback, PredictorState predictorState) throws InterruptedException, ExecutionException

	{
		PredictiveUnitState rootState = predictorState.rootState;
		rootState.predictiveUnitBean.sendFeedback(feedback, rootState);
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
		
		PredictiveUnitState rootState = new PredictiveUnitState(rootUnit,containersMap,this.nodeTypeMap,this.nodeImplementationMap);
		
		return new PredictorState(rootUnit.getName(),rootState, enabled);
	
	}
	
	
}