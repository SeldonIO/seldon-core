package io.seldon.engine.predictors;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;


@Component
public class PredictorBean {

    private final Map<String,PredictiveUnitBean> nodeClassMap;
	
    @Autowired
	public PredictorBean(
			ModelUnit modelUnit, 
			RouterUnit routerUnit, 
			CombinerUnit combinerUnit, 
			SimpleModelUnit simpleModelUnit, 
			SimpleRouterUnit simpleRouterUnit,
			AverageCombinerUnit averageCombinerUnit,
			RandomABTestUnit randomABTestUnit) {
        nodeClassMap = new HashMap<String,PredictiveUnitBean>();
    	nodeClassMap.put(PredictiveUnitDef.PredictiveUnitType.MODEL.toString() + "_" + PredictiveUnitDef.PredictiveUnitSubType.MICROSERVICE.toString(), modelUnit);
    	nodeClassMap.put(PredictiveUnitDef.PredictiveUnitType.MODEL.toString() + "_" + PredictiveUnitDef.PredictiveUnitSubType.SIMPLE_MODEL.toString(), simpleModelUnit);
    	nodeClassMap.put(PredictiveUnitDef.PredictiveUnitType.ROUTER.toString() + "_" + PredictiveUnitDef.PredictiveUnitSubType.MICROSERVICE.toString(), routerUnit);
    	nodeClassMap.put(PredictiveUnitDef.PredictiveUnitType.ROUTER.toString() + "_" + PredictiveUnitDef.PredictiveUnitSubType.SIMPLE_ROUTER.toString(), simpleRouterUnit);
    	nodeClassMap.put(PredictiveUnitDef.PredictiveUnitType.COMBINER.toString() + "_" + PredictiveUnitDef.PredictiveUnitSubType.MICROSERVICE.toString(), combinerUnit);
    	nodeClassMap.put(PredictiveUnitDef.PredictiveUnitType.COMBINER.toString() + "_" + PredictiveUnitDef.PredictiveUnitSubType.AVERAGE_COMBINER.toString(), averageCombinerUnit);
    	nodeClassMap.put(PredictiveUnitDef.PredictiveUnitType.ROUTER.toString() + "_" + PredictiveUnitDef.PredictiveUnitSubType.RANDOM_ABTEST.toString(), randomABTestUnit);
    }
   
	
	public PredictionResponseDef predict(PredictionRequestDef request, PredictorState predictorState) throws InterruptedException, ExecutionException
	{
		PredictiveUnitState rootState = predictorState.rootState;
		Future<PredictionResponseDef> ret = rootState.predictiveUnitBean.predict(request, rootState);
		return ret.get();
	}
	
	//TODO
	public PredictorState predictorStateFromDeploymentDef(PredictorDef predictorDef){
		String rootId = predictorDef.getRoot();
        Boolean enabled = predictorDef.getEnabled();
        List<PredictiveUnitDef> predictiveUnitDefList = predictorDef.getPredictiveUnitsList();
        
        Map<String,PredictiveUnitDef> predictiveUnitDefMap = new HashMap<>();
        Map<String,PredictiveUnitState> predictiveUnitStateMap = new HashMap<>();
     
        // First we go through all the nodes, instantiate the PredictorNode objects and populate dictionaries
        for (PredictiveUnitDef predictiveUnitDef : predictiveUnitDefList){
        	
        	String id = predictiveUnitDef.getId();
            
        	String typeSubtype = predictiveUnitDef.getType() + "_" + predictiveUnitDef.getSubtype();
        	PredictiveUnitBean predictiveUnitBean = nodeClassMap.get(typeSubtype);
        	
        	PredictiveUnitState predictiveUnitState = new PredictiveUnitState(predictiveUnitDef);
        	
        	predictiveUnitState.setPredictiveUnitBean(predictiveUnitBean);
        	
        	predictiveUnitDefMap.put(id, predictiveUnitDef);
        	predictiveUnitStateMap.put(id, predictiveUnitState);
        }
        
        // Then we go through the json nodes again and add the children links
        for (Map.Entry<String, PredictiveUnitDef> entry : predictiveUnitDefMap.entrySet()) {
        	PredictiveUnitDef predictiveUnitDef = entry.getValue();
        	PredictiveUnitState predictiveUnitState = predictiveUnitStateMap.get(entry.getKey());
        	
        	List<String> childIds = predictiveUnitDef.getChildrenList();
       
        	for (String childId : childIds)
            {
        		predictiveUnitState.addChild(childId, predictiveUnitStateMap.get(childId));
            }
        }
        
        // TODO: When predicting, the predictor will be stuck in a loop if the json is malformed 
        // and the graph contains a cycle. Maybe add some code to check that the json is well formed
        
        return new PredictorState(rootId,predictiveUnitStateMap,enabled);
	}
	   
	public PredictorState predictorStateFromDeploymentDef(DeploymentDef deploymentDef){
		PredictorDef predictorDef = deploymentDef.getPredictor();
		
        return predictorStateFromDeploymentDef(predictorDef);
	}
	
	
}