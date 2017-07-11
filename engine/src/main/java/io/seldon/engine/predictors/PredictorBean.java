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
    	nodeClassMap.put("model_external", modelUnit);
    	nodeClassMap.put("model_simpleModel", simpleModelUnit);
    	nodeClassMap.put("router_external", routerUnit);
    	nodeClassMap.put("router_simpleRouter", simpleRouterUnit);
    	nodeClassMap.put("combiner_external", combinerUnit);
    	nodeClassMap.put("combiner_averageCombiner", averageCombinerUnit);
    	nodeClassMap.put("router_randomABTest", randomABTestUnit);
    }
   
	
	public PredictorReturn predict(PredictorRequest request, PredictorState predictorState) throws InterruptedException, ExecutionException
	{
		PredictiveUnitState rootState = predictorState.rootState;
		Future<PredictorReturn> ret = rootState.predictiveUnitBean.predict(request, rootState);
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