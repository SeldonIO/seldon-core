package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.ArrayList;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Response;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Request;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.engine.exception.APIException;


@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected Response backwardPass(List<Response> inputs, Request request, PredictiveUnitState state){
		return inputs.get(0);
	}
	
	@Override
	public List<PredictiveUnitState> forwardPass(Request request, PredictiveUnitState state, Map<String,Integer> routingDict){
		Integer branchIndex = forwardPass(request, state);
		boolean  isPossible = sanityCheckRouting(branchIndex, state);
		if (!isPossible){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,"Router that caused the exception: id="+state.name+" name="+state.name);
		}
		populateRoutingDict(branchIndex, routingDict, state);
		
		List<PredictiveUnitState> route = new ArrayList<PredictiveUnitState>();
		route.add(state.children.get(branchIndex));
		return route;
		
	}
	
	@Override
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		internalPredictionService.sendFeedbackRouter(feedback, state.endpoint);
	}
	
	protected Integer forwardPass(Request request, PredictiveUnitState state){
		Response ret = internalPredictionService.getRouting(request, state.endpoint);
		int branchIndex = (int) ret.getData().getTensor().getValues(0);
		return branchIndex;
	}
	
	private void populateRoutingDict(Integer branchIndex, Map<String, Integer> routingDict, PredictiveUnitState state){
		routingDict.put(state.name, branchIndex);
	}
	
	private boolean sanityCheckRouting(Integer branchIndex, PredictiveUnitState state){
		return branchIndex < state.children.size();
	}

}
