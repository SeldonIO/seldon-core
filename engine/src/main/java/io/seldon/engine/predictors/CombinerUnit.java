package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.engine.service.PredictionServiceRequest;

@Component
public class CombinerUnit extends PredictiveUnitBean{
	
	public CombinerUnit() {
		super();
	}

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictionServiceRequest request, PredictiveUnitState state){
		return new ArrayList<PredictiveUnitState>(state.children.values());
	}

}
