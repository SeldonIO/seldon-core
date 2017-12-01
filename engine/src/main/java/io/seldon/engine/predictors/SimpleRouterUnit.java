package io.seldon.engine.predictors;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@Component
public class SimpleRouterUnit extends RouterUnit {

    public SimpleRouterUnit() {}

	@Override
	protected int route(SeldonMessage input, PredictiveUnitState state){
		return 0;
	} 
	
	@Override
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		return;
	}

}
