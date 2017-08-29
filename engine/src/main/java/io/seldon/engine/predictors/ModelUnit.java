package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionResponseMetaDef;

@Component
public class ModelUnit extends PredictiveUnitBean{
	
	public ModelUnit() {
		super();
	}
	
	private PredictionResponseDef doPredict(PredictionRequestDef request, PredictiveUnitState state)
	{
		PredictionResponseDef ret = null;
		try {
			ret = internalPredictionService.getPrediction(request, state.endpoint);
//			ret = PredictionResponseDef.newBuilder(ret).setMeta(PredictionResponseMetaDef.newBuilder(ret.getMeta()).addModel(state.id)).build();
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return ret;
	}

	@Override
	@Async
	protected Future<PredictionResponseDef> predict(PredictionRequestDef request, PredictiveUnitState state, Map<String,Integer> routingDict) throws InterruptedException, ExecutionException{
		return new AsyncResult<>(doPredict(request,state));
	}

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictionRequestDef request, PredictiveUnitState state, Map<String,Integer> routingDict){
		return new ArrayList<PredictiveUnitState>(Arrays.asList(state));
	}
	
	
}
