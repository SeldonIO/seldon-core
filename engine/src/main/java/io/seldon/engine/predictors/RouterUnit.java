package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.ArrayList;

import org.nd4j.linalg.api.ndarray.INDArray;
import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.predictors.PredictorUtils;


@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected Message backwardPass(List<Message> inputs, Message request, PredictiveUnitState state){
		return inputs.get(0);
	}
	
	@Override
	public List<PredictiveUnitState> forwardPass(Message request, PredictiveUnitState state, Map<String,Integer> routingDict){
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
	
	protected Integer forwardPass(Message request, PredictiveUnitState state){
		Message ret = internalPredictionService.getRouting(request, state.endpoint);
		int branchIndex = getBranchIndex(ret,state);
		return branchIndex;
	}
	
	private void populateRoutingDict(Integer branchIndex, Map<String, Integer> routingDict, PredictiveUnitState state){
		routingDict.put(state.name, branchIndex);
	}
	
	private int getBranchIndex(Message routerReturn, PredictiveUnitState state){
		int branchIndex = 0;
		try {
			INDArray dataArray = PredictorUtils.getINDArray(routerReturn.getData());
			branchIndex = (int) dataArray.getInt(0);
			// branchIndex = (int) routerReturn.getData().getTensor().getValues(0);
		}
		catch (IndexOutOfBoundsException e) {
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,"Router that caused the exception: id="+state.name+" name="+state.name);
		}
		return branchIndex;
	}
	
	private boolean sanityCheckRouting(Integer branchIndex, PredictiveUnitState state){
		return branchIndex < state.children.size();
	}

}
