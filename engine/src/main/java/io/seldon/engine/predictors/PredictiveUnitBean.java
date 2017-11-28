package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import io.seldon.engine.service.InternalPredictionService;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Meta;

@Component
public abstract class PredictiveUnitBean {

	@Autowired
	InternalPredictionService internalPredictionService;
	
	public PredictiveUnitBean() {}
	
	public PredictiveUnitBean(InternalPredictionService internalPredictionService){
		this.internalPredictionService = internalPredictionService;
	}
	
	public void sendFeedback(Feedback feedback, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		Future<Boolean> ret = sendFeedbackAsync(feedback,state);
		ret.get();
	}
	
	@Async
	public Future<Boolean> sendFeedbackAsync(Feedback feedback, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		System.out.println("NODE " + state.name + ": entered feedback");
		List<PredictiveUnitState> children = state.children;	
		List<Future<Boolean>> returns = new ArrayList<Future<Boolean>>();
		
		// First we call sendFeebackAsync on children
		for (PredictiveUnitState child : children){
			returns.add(child.predictiveUnitBean.sendFeedbackAsync(feedback,child));
		}
		
		// Then we wait for our own feedback
		if (feedback.getResponse().getMeta().getRoutingMap().get(state.name)!=null){
			// If the response routing dictionary contains the current predictive unit key
			doSendFeedback(feedback, state);
			PredictiveUnitState chosenRoute = state.children.get(feedback.getResponse().getMeta().getRoutingMap().get(state.name));
			chosenRoute.predictiveUnitBean.doStoreFeedbackMetrics(feedback, chosenRoute);
		}
		
		//Then we wait for children feedback
		for (Future<Boolean> ret : returns){
			ret.get();
		}
		
		return new AsyncResult<>(true);
	}
	
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		return;
	}
	
	protected void doStoreFeedbackMetrics(Feedback feedback, PredictiveUnitState state){
		return;
	}
	
	
	public Message predict(Message request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		Map<String,Integer> routingDict = new HashMap<String,Integer>();
		Future<Message> ret = state.predictiveUnitBean.predict(request, state, routingDict);
		Message response = ret.get();
		Message.Builder builder = Message
	    		.newBuilder(response)
	    		.setMeta(Meta
	    				.newBuilder(response.getMeta()).putAllRouting(routingDict));
		return builder.build();
	}
	
	@Async
	protected Future<Message> predict(Message request, PredictiveUnitState state, Map<String,Integer> routingDict) throws InterruptedException, ExecutionException{
		System.out.println("NODE " + state.name + ": entered predict");
		List<PredictiveUnitState> routing = forwardPass(request,state,routingDict);
		System.out.println("NODE " + state.name + ": got routing");
		
		List<Message> inputs = new ArrayList<>();
		
		List<Future<Message>> futureInputs = new ArrayList<>();
			
		for (PredictiveUnitState route : routing)
		{
			PredictiveUnitBean childBean = route.predictiveUnitBean;
			futureInputs.add(childBean.predict(request,route, routingDict));
		}
		System.out.println("NODE " + state.name + ": called child futures");
		
		for (Future<Message> futureInput : futureInputs)
		{
			inputs.add(futureInput.get());
		}
		System.out.println("NODE " + state.name + ": got child futures");
		
		System.out.println("NODE " + state.name + ": returning");
		return new AsyncResult<>(backwardPass(inputs,request,state));
	}
	
	protected List<PredictiveUnitState> forwardPass(Message request, PredictiveUnitState data, Map<String,Integer> routingDict){
		return null;
	}
	
	protected Message backwardPass(List<Message> inputs, Message request, PredictiveUnitState data){
		return null;
	}
}
