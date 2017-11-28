package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Message;

@Component
public class SimpleRouterUnit extends RouterUnit {

    public SimpleRouterUnit() {}

	@Override
	protected Integer forwardPass(Message request, PredictiveUnitState state){
		return 0;
	} 
	
	@Override
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		return;
	}

}
