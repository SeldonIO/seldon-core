package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import com.google.protobuf.Descriptors.FieldDescriptor;

import io.seldon.engine.service.InternalPredictionService;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionResponseMetaDef;
import io.seldon.protos.PredictionProtos.PredictionFeedbackDef;

@Component
public abstract class PredictiveUnitBean {

	@Autowired
	InternalPredictionService internalPredictionService;
	
	public PredictiveUnitBean() {}
	
	public PredictiveUnitBean(InternalPredictionService internalPredictionService){
		this.internalPredictionService = internalPredictionService;
	}
	
	@Async
	public void sendFeedback(PredictionFeedbackDef feedback, PredictiveUnitState state){
		System.out.println("NODE " + state.name + ": entered feedback");
		List<PredictiveUnitState> children = state.children;
		if (feedback.getResponse().getMeta().getRoutingMap().get(state.id)!=null){
			// If the response routing dictionary contains the current predictive unit key
			doSendFeedback(feedback, state);
		}
		for (PredictiveUnitState child : children){
			sendFeedback(feedback,child);
		}
	}
	
	protected void doSendFeedback(PredictionFeedbackDef feedback, PredictiveUnitState state){
		return;
	}
	
	public PredictionResponseDef predict(PredictionRequestDef request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		Map<String,Integer> routingDict = new HashMap<String,Integer>();
		Future<PredictionResponseDef> ret = state.predictiveUnitBean.predict(request, state, routingDict);
		PredictionResponseDef response = ret.get();
//		PredictionResponseMetaDef meta = response.getMeta();
//		Map<String, Integer> currentRouting = meta.getRoutingMap();
//		Iterator<Entry<String,Integer>> it = routingDict.entrySet().iterator();
//	    while (it.hasNext()) {
//	        Entry<String,Integer> pair = (Entry<String,Integer>)it.next();
//	        currentRouting.put(pair.getKey(), pair.getValue());
//	        it.remove(); // avoids a ConcurrentModificationException
//	    }
//	    FieldDescriptor field = PredictionResponseMetaDef.getDescriptor().findFieldByNumber(PredictionResponseMetaDef.ROUTING_FIELD_NUMBER);
		PredictionResponseDef.Builder builder = PredictionResponseDef
	    		.newBuilder(response)
	    		.setMeta(PredictionResponseMetaDef
	    				.newBuilder(response.getMeta()).putAllRouting(routingDict));
		return builder.build();
	}
	
	@Async
	protected Future<PredictionResponseDef> predict(PredictionRequestDef request, PredictiveUnitState state, Map<String,Integer> routingDict) throws InterruptedException, ExecutionException{
		System.out.println("NODE " + state.name + ": entered predict");
		List<PredictiveUnitState> routing = forwardPass(request,state,routingDict);
		System.out.println("NODE " + state.name + ": got routing");
		
		List<PredictionResponseDef> inputs = new ArrayList<>();
		
		List<Future<PredictionResponseDef>> futureInputs = new ArrayList<>();
			
		for (PredictiveUnitState route : routing)
		{
			PredictiveUnitBean childBean = route.predictiveUnitBean;
			futureInputs.add(childBean.predict(request,route, routingDict));
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
	
	protected List<PredictiveUnitState> forwardPass(PredictionRequestDef request, PredictiveUnitState data, Map<String,Integer> routingDict){
		return null;
	}
	
	protected PredictionResponseDef backwardPass(List<PredictionResponseDef> inputs, PredictiveUnitState data){
		return null;
	}
}
