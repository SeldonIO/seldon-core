package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitMethod;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;

public class PredictorConfigBean {
    public final Map<PredictiveUnitType,List<PredictiveUnitMethod>> typeMethodsMap;
    public final Map<PredictiveUnitImplementation,PredictiveUnitBean> nodeImplementationMap;
    
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
        typeMethodsMap.put(PredictiveUnitType.COMBINER,combinerMethods);
        
        // -------------------------
        // HARDCODED IMPLEMENTATIONS
        // -------------------------
        nodeImplementationMap = new HashMap<PredictiveUnitImplementation,PredictiveUnitBean>();
        nodeImplementationMap.put(PredictiveUnitImplementation.AVERAGE_COMBINER, averageCombinerUnit);
        nodeImplementationMap.put(PredictiveUnitImplementation.SIMPLE_MODEL, simpleModelUnit);
        nodeImplementationMap.put(PredictiveUnitImplementation.SIMPLE_ROUTER, simpleRouterUnit);
        nodeImplementationMap.put(PredictiveUnitImplementation.RANDOM_ABTEST, randomABTestUnit);
    }
    
    public PredictiveUnitBean getImplementation(PredictiveUnitState state){
    	if (state.implementation != PredictiveUnitImplementation.UNKNOWN_IMPLEMENTATION){
    		return nodeImplementationMap.get(state.implementation);
    	}
    	return null;
    }
    
    public Boolean hasMethod(PredictiveUnitMethod method, PredictiveUnitState state){
    	if (state.implementation != PredictiveUnitImplementation.UNKNOWN_IMPLEMENTATION){
    		return false;
    	}
    	if (state.type == PredictiveUnitType.UNKNOWN_TYPE) {
    		return state.methods.contains(method);
    	}
    	else
    	{
    		return typeMethodsMap.get(state.type).contains(method);
    	}
    }
}
