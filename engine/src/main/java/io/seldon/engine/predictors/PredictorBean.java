package io.seldon.engine.predictors;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitSubtype;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.PredictionProtos.FeedbackDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.ResponseDef;

import io.kubernetes.client.proto.V1.Container;


@Component
public class PredictorBean {

    public final Map<PredictiveUnitType,Map<PredictiveUnitSubtype,PredictiveUnitBean>> nodeClassMap;
	
    @Autowired
	public PredictorBean(
			ModelUnit modelUnit, 
			RouterUnit routerUnit, 
			CombinerUnit combinerUnit, 
			SimpleModelUnit simpleModelUnit, 
			SimpleRouterUnit simpleRouterUnit,
			AverageCombinerUnit averageCombinerUnit,
			RandomABTestUnit randomABTestUnit) {
        nodeClassMap = new HashMap<PredictiveUnitType,Map<PredictiveUnitSubtype,PredictiveUnitBean>>();
        
        Map<PredictiveUnitSubtype,PredictiveUnitBean> modelsMap = new HashMap<PredictiveUnitSubtype,PredictiveUnitBean>();
        modelsMap.put(PredictiveUnitSubtype.MICROSERVICE, modelUnit);
        modelsMap.put(PredictiveUnitSubtype.SIMPLE_MODEL, simpleModelUnit);
        nodeClassMap.put(PredictiveUnitType.MODEL, modelsMap);
        
        Map<PredictiveUnitSubtype,PredictiveUnitBean> routersMap = new HashMap<PredictiveUnitSubtype,PredictiveUnitBean>();
        routersMap.put(PredictiveUnitSubtype.MICROSERVICE, routerUnit);
        routersMap.put(PredictiveUnitSubtype.RANDOM_ABTEST, randomABTestUnit);
        routersMap.put(PredictiveUnitSubtype.SIMPLE_ROUTER, simpleRouterUnit);
        nodeClassMap.put(PredictiveUnitType.ROUTER, routersMap);
        
        Map<PredictiveUnitSubtype,PredictiveUnitBean> combinersMap = new HashMap<PredictiveUnitSubtype,PredictiveUnitBean>();
        combinersMap.put(PredictiveUnitSubtype.MICROSERVICE, combinerUnit);
        combinersMap.put(PredictiveUnitSubtype.AVERAGE_COMBINER, averageCombinerUnit);
        nodeClassMap.put(PredictiveUnitType.COMBINER, combinersMap);
        
    }
   
	public ResponseDef predict(RequestDef request, PredictorState predictorState) throws InterruptedException, ExecutionException

	{
		PredictiveUnitState rootState = predictorState.rootState;
		return rootState.predictiveUnitBean.predict(request, rootState);
	}
	
	public void sendFeedback(FeedbackDef feedback, PredictorState predictorState) throws InterruptedException, ExecutionException

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
		
		PredictiveUnitState rootState = new PredictiveUnitState(rootUnit,containersMap,this.nodeClassMap);
		
		return new PredictorState(rootUnit.getName(),rootState, enabled);
	
	}
	
	
}