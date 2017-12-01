package io.seldon.engine.predictors;

import org.nd4j.linalg.api.ndarray.INDArray;
import org.springframework.stereotype.Component;

import com.google.protobuf.InvalidProtocolBufferException;

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
		Message ret = null;
		try {
			ret = internalPredictionService.route(input, state);
		} catch (InvalidProtocolBufferException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		int branchIndex = getBranchIndex(ret,state);
		return branchIndex;
	}
	
	@Override
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		try {
			internalPredictionService.sendFeedback(feedback, state);
		} catch (InvalidProtocolBufferException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
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
