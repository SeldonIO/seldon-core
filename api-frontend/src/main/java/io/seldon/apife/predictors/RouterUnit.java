package io.seldon.apife.predictors;

import java.util.List;

import org.springframework.stereotype.Component;

@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected PredictorReturn backwardPass(List<PredictorData> inputs, PredictiveUnitState state){
		return (PredictorReturn) inputs.get(0);
	}

}
