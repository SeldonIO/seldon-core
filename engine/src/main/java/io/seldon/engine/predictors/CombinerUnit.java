package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.RequestDef;

@Component
public class CombinerUnit extends PredictiveUnitBean{
	
	public CombinerUnit() {
		super();
	}

	@Override
	protected List<PredictiveUnitState> forwardPass(RequestDef request, PredictiveUnitState state, Map<String,Integer> routingDict){
		return state.children;
	}

}
