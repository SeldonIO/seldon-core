package io.seldon.engine.predictors;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@Component
public class SimpleRouterUnit extends PredictiveUnitBean {

    public SimpleRouterUnit() {}

	@Override
	public int route(SeldonMessage input, PredictiveUnitState state){
		return 0;
	}
}
