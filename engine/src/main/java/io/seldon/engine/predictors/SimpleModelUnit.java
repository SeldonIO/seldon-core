package io.seldon.engine.predictors;

import java.util.Arrays;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.MetaDef;
import io.seldon.protos.PredictionProtos.ResponseDef;
import io.seldon.protos.PredictionProtos.MetaDef;
import io.seldon.protos.PredictionProtos.StatusDef;
import io.seldon.protos.PredictionProtos.Tensor;

@Component
public class SimpleModelUnit extends ModelUnit {
	
	public SimpleModelUnit() {}

	public static final Double[] values = {0.1,0.9,0.5};		
	public static final String[] classes = {"class0","class1","class2"};
	
	private ResponseDef doPredict(RequestDef request, PredictiveUnitState state)
	{
		ResponseDef ret = ResponseDef.newBuilder()
				.setStatus(StatusDef.newBuilder().setStatus(StatusDef.Status.SUCCESS).build())
				.setMeta(MetaDef.newBuilder())//.addModel(state.id))
				.setData(DefaultDataDef.newBuilder().addAllNames(Arrays.asList(classes))
					.setTensor(Tensor.newBuilder().addShape(1).addShape(values.length)
					.addAllValues(Arrays.asList(values)))).build();
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
	protected Future<ResponseDef> predict(RequestDef request, PredictiveUnitState state, Map<String, Integer> routingDict) throws InterruptedException, ExecutionException{
		System.out.println("Model " + state.name + " starting computations");
		
		return new AsyncResult<>(doPredict(request,state));
	}
	
	
}
