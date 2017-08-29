package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.PredictionFeedbackDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;

@Component
public class SimpleRouterUnit extends RouterUnit {

    public SimpleRouterUnit() {}

	@Override
	protected Integer forwardPass(PredictionRequestDef request, PredictiveUnitState state){
		return 0;
	} 
	
	@Override
	protected void doSendFeedback(PredictionFeedbackDef feedback, PredictiveUnitState state){
		return;
	}

}
