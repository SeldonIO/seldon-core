package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Metrics;
import io.micrometer.core.instrument.Tag;
import io.seldon.engine.service.InternalPredictionService;
import io.seldon.protos.PredictionProtos.FeedbackDef;
import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.ResponseDef;
import io.seldon.protos.PredictionProtos.MetaDef;

@Component
public abstract class PredictiveUnitBean {

	@Autowired
	InternalPredictionService internalPredictionService;
	
	public PredictiveUnitBean() {}
	
	public PredictiveUnitBean(InternalPredictionService internalPredictionService){
		this.internalPredictionService = internalPredictionService;
	}
	
	@Async
	public void sendFeedback(FeedbackDef feedback, PredictiveUnitState state){
		System.out.println("NODE " + state.name + ": entered feedback");
		List<PredictiveUnitState> children = state.children;
		if (feedback.getResponse().getMeta().getRoutingMap().get(state.id)!=null){
			// If the response routing dictionary contains the current predictive unit key
			doSendFeedback(feedback, state);
			PredictiveUnitState chosenRoute = state.children.get(feedback.getResponse().getMeta().getRoutingMap().get(state.id));
			chosenRoute.predictiveUnitBean.doStoreFeedbackMetrics(feedback, chosenRoute);
		}
		for (PredictiveUnitState child : children){
			child.predictiveUnitBean.sendFeedback(feedback,child);
		}
	}
	
	protected void doSendFeedback(FeedbackDef feedback, PredictiveUnitState state){
		return;
	}
	
	protected void doStoreFeedbackMetrics(FeedbackDef feedback, PredictiveUnitState state){
		return;
	}
	
	
	public ResponseDef predict(RequestDef request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		Map<String,Integer> routingDict = new HashMap<String,Integer>();
		Future<ResponseDef> ret = state.predictiveUnitBean.predict(request, state, routingDict);
		ResponseDef response = ret.get();
//		MetaDef meta = response.getMeta();
//		Map<String, Integer> currentRouting = meta.getRoutingMap();
//		Iterator<Entry<String,Integer>> it = routingDict.entrySet().iterator();
//	    while (it.hasNext()) {
//	        Entry<String,Integer> pair = (Entry<String,Integer>)it.next();
//	        currentRouting.put(pair.getKey(), pair.getValue());
//	        it.remove(); // avoids a ConcurrentModificationException
//	    }
//	    FieldDescriptor field = MetaDef.getDescriptor().findFieldByNumber(MetaDef.ROUTING_FIELD_NUMBER);
		ResponseDef.Builder builder = ResponseDef
	    		.newBuilder(response)
	    		.setMeta(MetaDef
	    				.newBuilder(response.getMeta()).putAllRouting(routingDict));
		return builder.build();
	}
	
	@Async
	protected Future<ResponseDef> predict(RequestDef request, PredictiveUnitState state, Map<String,Integer> routingDict) throws InterruptedException, ExecutionException{
		System.out.println("NODE " + state.name + ": entered predict");
		List<PredictiveUnitState> routing = forwardPass(request,state,routingDict);
		System.out.println("NODE " + state.name + ": got routing");
		
		List<ResponseDef> inputs = new ArrayList<>();
		
		List<Future<ResponseDef>> futureInputs = new ArrayList<>();
			
		for (PredictiveUnitState route : routing)
		{
			PredictiveUnitBean childBean = route.predictiveUnitBean;
			futureInputs.add(childBean.predict(request,route, routingDict));
		}
		System.out.println("NODE " + state.name + ": called child futures");
		
		for (Future<ResponseDef> futureInput : futureInputs)
		{
			inputs.add(futureInput.get());
		}
		System.out.println("NODE " + state.name + ": got child futures");
		
		System.out.println("NODE " + state.name + ": returning");
		return new AsyncResult<>(backwardPass(inputs,request,state));
	}
	
	protected List<PredictiveUnitState> forwardPass(RequestDef request, PredictiveUnitState data, Map<String,Integer> routingDict){
		return null;
	}
	
	protected ResponseDef backwardPass(List<ResponseDef> inputs, RequestDef request, PredictiveUnitState data){
		return null;
	}
}
