package io.seldon.engine.predictors;

import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.PredictionResponseDef;

@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected PredictionResponseDef backwardPass(List<PredictionResponseDef> inputs, PredictiveUnitState state){
		return inputs.get(0);
	}

}
