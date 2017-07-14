package io.seldon.engine.predictors;

import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.seldon.engine.service.PredictionServiceRequest;

@Component
public class SimpleModelUnit extends ModelUnit {
	
	public SimpleModelUnit() {}

	public static final Double[][] values = {{0.1,0.9,0.5}};		
	public static final String[] classes = {"class0","class1","class2"};
	
	private PredictorReturn doPredict(PredictionServiceRequest request, PredictiveUnitState state)
	{
		
		PredictorReturn ret = new PredictorReturn(classes, values);
		try {
			Thread.sleep(20);
		} catch (InterruptedException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		System.out.println("Model " + state.name + " finishing computations");
		return ret;
	}
	
	@Override
	public Future<PredictorReturn> predict(PredictionServiceRequest request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		System.out.println("Model " + state.name + " starting computations");
		
		return new AsyncResult<>(doPredict(request,state));
	}
	
	
}
