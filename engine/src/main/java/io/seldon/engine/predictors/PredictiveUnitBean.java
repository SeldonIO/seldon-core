package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.seldon.engine.service.InternalPredictionService;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;

@Component
public abstract class PredictiveUnitBean {

	@Autowired
	InternalPredictionService internalPredictionService;
	
	public PredictiveUnitBean() {}
	
	public PredictiveUnitBean(InternalPredictionService internalPredictionService){
		this.internalPredictionService = internalPredictionService;
	}
	
	@Async
	public Future<PredictionResponseDef> predict(PredictionRequestDef request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		System.out.println("NODE " + state.name + ": entered predict");
		List<PredictiveUnitState> routing = forwardPass(request,state);
		System.out.println("NODE " + state.name + ": got routing");
		
		List<PredictionResponseDef> inputs = new ArrayList<>();
		
		List<Future<PredictionResponseDef>> futureInputs = new ArrayList<>();
			
		for (PredictiveUnitState route : routing)
		{
			PredictiveUnitBean childBean = route.predictiveUnitBean;
			futureInputs.add(childBean.predict(request,route));
		}
		System.out.println("NODE " + state.name + ": called child futures");
		
		for (Future<PredictionResponseDef> futureInput : futureInputs)
		{
			inputs.add(futureInput.get());
		}
		System.out.println("NODE " + state.name + ": got child futures");
		
		System.out.println("NODE " + state.name + ": returning");
		return new AsyncResult<>(backwardPass(inputs,state));
	}
	
	protected List<PredictiveUnitState> forwardPass(PredictionRequestDef request, PredictiveUnitState data){
		return null;
	}
	
	protected PredictionResponseDef backwardPass(List<PredictionResponseDef> inputs, PredictiveUnitState data){
		return null;
	}
}
