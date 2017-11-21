package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.FeedbackDef;
import io.seldon.protos.PredictionProtos.RequestDef;

@Component
public class SimpleRouterUnit extends RouterUnit {

    public SimpleRouterUnit() {}

	@Override
	protected Integer forwardPass(RequestDef request, PredictiveUnitState state){
		return 0;
	} 
	
	@Override
	protected void doSendFeedback(FeedbackDef feedback, PredictiveUnitState state){
		return;
	}

}
