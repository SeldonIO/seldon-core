package io.seldon.engine.predictors;

import java.util.Arrays;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.DefaultDataValues;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionStatusDef;

@Component
public class SimpleModelUnit extends ModelUnit {
	
	public SimpleModelUnit() {}

	public static final Double[] values = {0.1,0.9,0.5};		
	public static final String[] classes = {"class0","class1","class2"};
	
	private PredictionResponseDef doPredict(PredictionRequestDef request, PredictiveUnitState state)
	{
		
		PredictionResponseDef ret = PredictionResponseDef.newBuilder().setStatus(PredictionStatusDef.newBuilder().setStatus(PredictionStatusDef.Status.SUCCESS).build())
			.setResponse(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(classes))
					.addValues(DefaultDataValues.newBuilder().addAllValue(Arrays.asList(values)))
					.build()).build();
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
	public Future<PredictionResponseDef> predict(PredictionRequestDef request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		System.out.println("Model " + state.name + " starting computations");
		
		return new AsyncResult<>(doPredict(request,state));
	}
	
	
}
