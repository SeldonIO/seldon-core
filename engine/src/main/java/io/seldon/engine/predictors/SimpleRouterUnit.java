package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import org.springframework.stereotype.Component;

@Component
public class SimpleRouterUnit extends RouterUnit {

    public SimpleRouterUnit() {}

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictorData request, PredictiveUnitState state){
		List<PredictiveUnitState> ret = new ArrayList<>();
		ret.add(state.children.get(state.children.keySet().toArray()[0]));
		return ret;
	} 

}
