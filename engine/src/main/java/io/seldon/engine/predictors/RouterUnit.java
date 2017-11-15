package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.ArrayList;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.ResponseDef;
import io.seldon.protos.PredictionProtos.FeedbackDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.MetaDef;
import io.seldon.engine.exception.APIException;


@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected ResponseDef backwardPass(List<ResponseDef> inputs, RequestDef request, PredictiveUnitState state){
		return inputs.get(0);
	}
	
	@Override
	public List<PredictiveUnitState> forwardPass(RequestDef request, PredictiveUnitState state, Map<String,Integer> routingDict){
		Integer branchIndex = forwardPass(request, state);
		boolean  isPossible = sanityCheckRouting(branchIndex, state);
		if (!isPossible){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,"Router that caused the exception: id="+state.id+" name="+state.name);
		}
		populateRoutingDict(branchIndex, routingDict, state);
		
		List<PredictiveUnitState> route = new ArrayList<PredictiveUnitState>();
		route.add(state.children.get(branchIndex));
		return route;
		
	}
	
	@Override
	protected void doSendFeedback(FeedbackDef feedback, PredictiveUnitState state){
		internalPredictionService.sendFeedbackRouter(feedback, state.endpoint);
	}
	
	protected Integer forwardPass(RequestDef request, PredictiveUnitState state){
		ResponseDef ret = internalPredictionService.getRouting(request, state.endpoint);
		int branchIndex = (int) ret.getData().getTensor().getValues(0);
		return branchIndex;
	}
	
	private void populateRoutingDict(Integer branchIndex, Map<String, Integer> routingDict, PredictiveUnitState state){
		routingDict.put(state.id, branchIndex);
	}
	
	private boolean sanityCheckRouting(Integer branchIndex, PredictiveUnitState state){
		return branchIndex < state.children.size();
	}

}
