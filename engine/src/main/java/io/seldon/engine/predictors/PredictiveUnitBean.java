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

@Component
public abstract class PredictiveUnitBean {

	@Autowired
	InternalPredictionService internalPredictionService;
	
	public PredictiveUnitBean() {}
	
	public PredictiveUnitBean(InternalPredictionService internalPredictionService){
		this.internalPredictionService = internalPredictionService;
	}
	
	@Async
	public Future<PredictorReturn> predict(PredictorData request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		System.out.println("NODE " + state.name + ": entered predict");
		List<PredictiveUnitState> routing = forwardPass(request,state);
		System.out.println("NODE " + state.name + ": got routing");
		
		List<PredictorData> inputs = new ArrayList<>();
		if (routing != null){
			
			List<Future<PredictorReturn>> futureInputs = new ArrayList<>();
			
			for (PredictiveUnitState child : routing)
			{
				PredictiveUnitBean childBean = child.predictiveUnitBean;
				futureInputs.add(childBean.predict(request,child));
			}
			System.out.println("NODE " + state.name + ": called child futures");
			
			for (Future<PredictorReturn> futureInput : futureInputs)
			{
				inputs.add(futureInput.get());
			}
			System.out.println("NODE " + state.name + ": got child futures");
		}
		else{
			inputs.add(request);
		}
		
		System.out.println("NODE " + state.name + ": returning");
		return new AsyncResult<>(backwardPass(inputs,state));
	}
	
	protected List<PredictiveUnitState> forwardPass(PredictorData request, PredictiveUnitState data){
		return null;
	}
	
	protected PredictorReturn backwardPass(List<PredictorData> inputs, PredictiveUnitState data){
		return null;
	}
}
