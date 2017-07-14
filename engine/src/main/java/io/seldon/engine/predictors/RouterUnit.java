package io.seldon.engine.predictors;

import java.util.List;

import org.springframework.stereotype.Component;

@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected PredictorReturn backwardPass(List<PredictorReturn> inputs, PredictiveUnitState state){
		return (PredictorReturn) inputs.get(0);
	}

}
