package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.PredictionRequestDef;

@Component
public class SimpleRouterUnit extends RouterUnit {

    public SimpleRouterUnit() {}

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictionRequestDef request, PredictiveUnitState state){
		List<PredictiveUnitState> ret = new ArrayList<>();
		ret.add(state.children.get(0));
		return ret;
	} 

}
