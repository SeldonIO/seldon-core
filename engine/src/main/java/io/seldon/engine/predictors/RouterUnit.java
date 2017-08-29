package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.List;
import java.util.ArrayList;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionRequestMetaDef;
import io.seldon.protos.PredictionProtos.RouteResponseDef;

@Component
public class RouterUnit extends PredictiveUnitBean{
    
    public RouterUnit() {
    	super();
    }

	@Override
	protected PredictionResponseDef backwardPass(List<PredictionResponseDef> inputs, PredictiveUnitState state){
		return inputs.get(0);
	}
	
	@Override
	protected List<PredictiveUnitState> forwardPass(PredictionRequestDef request, PredictiveUnitState state){
		RouteResponseDef ret = internalPredictionService.getRouting(request, state.endpoint);
//			ret = PredictionResponseDef.newBuilder(ret).setMeta(PredictionResponseMetaDef.newBuilder(ret.getMeta()).addModel(state.id)).build();
		List<PredictiveUnitState> route = new ArrayList<PredictiveUnitState>();
		route.add(state.children.get(ret.getBranch()));
		return route;
	}

}
