package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.ArrayList;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionFeedbackDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionRequestMetaDef;
import io.seldon.protos.PredictionProtos.RouteResponseDef;

@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected PredictionResponseDef backwardPass(List<PredictionResponseDef> inputs, PredictiveUnitState state){
		return inputs.get(0);
	}
	
	@Override
	public List<PredictiveUnitState> forwardPass(PredictionRequestDef request, PredictiveUnitState state, Map<String,Integer> routingDict){
		Integer branchIndex = forwardPass(request, state);
		boolean  isPossible = sanityCheckRouting(branchIndex, state);
		if (!isPossible){
			//TODO: Add some sort of exception throwing
		}
		populateRoutingDict(branchIndex, routingDict, state);
		
		List<PredictiveUnitState> route = new ArrayList<PredictiveUnitState>();
		route.add(state.children.get(branchIndex));
		return route;
		
	}
	
	@Override
	protected void doSendFeedback(PredictionFeedbackDef feedback, PredictiveUnitState state){
		internalPredictionService.sendFeedback(feedback, state.endpoint);
	}
	
	protected Integer forwardPass(PredictionRequestDef request, PredictiveUnitState state){
		RouteResponseDef ret = internalPredictionService.getRouting(request, state.endpoint);
		return ret.getBranch();
	}
	
	private void populateRoutingDict(Integer branchIndex, Map<String, Integer> routingDict, PredictiveUnitState state){
		routingDict.put(state.id, branchIndex);
	}
	
	private boolean sanityCheckRouting(Integer branchIndex, PredictiveUnitState state){
		return branchIndex < state.children.size();
	}

}
