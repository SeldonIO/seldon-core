package io.seldon.engine.predictors;

import org.nd4j.linalg.api.ndarray.INDArray;
import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.predictors.PredictorUtils;


@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }
	
	@Override
	protected int route(Message input, PredictiveUnitState state){
		Message ret = internalPredictionService.getRouting(input, state.endpoint);
		int branchIndex = getBranchIndex(ret,state);
		return branchIndex;
	}
	
	@Override
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		internalPredictionService.sendFeedbackRouter(feedback, state.endpoint);
	}

	private int getBranchIndex(Message routerReturn, PredictiveUnitState state){
		int branchIndex = 0;
		try {
			INDArray dataArray = PredictorUtils.getINDArray(routerReturn.getData());
			branchIndex = (int) dataArray.getInt(0);
		}
		catch (IndexOutOfBoundsException e) {
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,"Router that caused the exception: id="+state.name+" name="+state.name);
		}
		return branchIndex;
	}

}
