package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.seldon.engine.service.PredictionServiceRequest;

@Component
public class ModelUnit extends PredictiveUnitBean{
	
	public ModelUnit() {
		super();
	}
	
	private PredictorReturn doPredict(PredictionServiceRequest request, PredictiveUnitState state)
	{
		PredictorReturn ret = null;
		try {
			ret = internalPredictionService.getPrediction(request, state.endpoint);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return ret;
	}

	@Override
	@Async
	public Future<PredictorReturn> predict(PredictionServiceRequest request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		return new AsyncResult<>(doPredict(request,state));
	}

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictionServiceRequest request, PredictiveUnitState state){
		return new ArrayList<PredictiveUnitState>(Arrays.asList(state));
	}
	
	
}
