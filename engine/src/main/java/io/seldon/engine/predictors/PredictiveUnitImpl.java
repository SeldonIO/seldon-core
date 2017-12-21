package io.seldon.engine.predictors;

import java.util.List;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

public abstract class PredictiveUnitImpl {

	public SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return input;
	}
	
	public SeldonMessage transformOutput(SeldonMessage output, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return output;
	}
	
	public SeldonMessage aggregate(List<SeldonMessage> outputs, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return outputs.get(0);
	}
	
	public int route(SeldonMessage input, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return -1;
	}
	
	public void doSendFeedback(Feedback feedback, PredictiveUnitState state) throws InvalidProtocolBufferException{
		return;
	}
}
