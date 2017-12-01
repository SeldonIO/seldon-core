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

import io.seldon.engine.exception.APIException;
import io.seldon.engine.service.InternalPredictionService;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Meta;

@Component
public abstract class PredictiveUnitBean {

	@Autowired
	InternalPredictionService internalPredictionService;
	
	public PredictiveUnitBean() {}
	
	public PredictiveUnitBean(InternalPredictionService internalPredictionService){
		this.internalPredictionService = internalPredictionService;
	}
	
	public SeldonMessage getOutput(SeldonMessage request, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		Map<String,Integer> routingDict = new HashMap<String,Integer>();
		SeldonMessage response = state.predictiveUnitBean.getOutputAsync(request, state, routingDict).get();
		SeldonMessage.Builder builder = SeldonMessage
	    		.newBuilder(response)
	    		.setMeta(Meta
	    				.newBuilder(response.getMeta()).putAllRouting(routingDict));
		return builder.build();
	}
	
	@Async
	protected Future<SeldonMessage> getOutputAsync(SeldonMessage input, PredictiveUnitState state, Map<String,Integer> routingDict) throws InterruptedException, ExecutionException{
		
		// Compute the transformed Input
		SeldonMessage transformedInput = state.predictiveUnitBean.transformInput(input, state);
		if (state.children.isEmpty()){
			// If this unit has no children, the transformed input becomes the output
			return new AsyncResult<>(transformedInput);
		}
		
		List<PredictiveUnitState> selectedChildren = new ArrayList<PredictiveUnitState>();
		List<Future<SeldonMessage>> deferredChildrenOutputs = new ArrayList<Future<SeldonMessage>>();
		List<SeldonMessage> childrenOutputs = new ArrayList<SeldonMessage>();
		
		// Get the routing. -1 means all children
		int routing = state.predictiveUnitBean.route(transformedInput, state);
		sanityCheckRouting(routing, state);
		// Update the routing dictionary
		routingDict.put(state.name, routing);
		
		if (routing == -1){
			// No routing, propagate to all children
			selectedChildren = state.children;
		}
		else
		{
			// Propagate to selected child only
			selectedChildren.add(state.children.get(routing));
		}
		
		// Get all the children outputs asynchronously
		for (PredictiveUnitState childState : selectedChildren){
			deferredChildrenOutputs.add(childState.predictiveUnitBean.getOutputAsync(transformedInput,childState,routingDict));
		}
		for (Future<SeldonMessage> deferredOutput : deferredChildrenOutputs){
			childrenOutputs.add(deferredOutput.get());
		}
		
		// Compute the backward transformation of all children outputs
		SeldonMessage aggregatedOutput = state.predictiveUnitBean.aggregateOutputs(childrenOutputs, state);
		SeldonMessage transformedOutput = state.predictiveUnitBean.transformOutput(aggregatedOutput, state);
		
		return new AsyncResult<>(transformedOutput);
		
	}
	
	public void sendFeedback(Feedback feedback, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		sendFeedbackAsync(feedback,state).get();
	}
	
	@Async
	public Future<Boolean> sendFeedbackAsync(Feedback feedback, PredictiveUnitState state) throws InterruptedException, ExecutionException{
		System.out.println("NODE " + state.name + ": entered feedback");
		List<PredictiveUnitState> children = new ArrayList<PredictiveUnitState>();
		List<Future<Boolean>> returns = new ArrayList<Future<Boolean>>();
		
		// First we determine children we will send feedback to according to routingDict info
		int routing = feedback.getResponse().getMeta().getRoutingMap().getOrDefault(state.name, -1);
		
		// TODO: Throw exception if routing is invalid (<-1 or > n_children)
		if (routing == -1){
			children = state.children;
		}
		else if (routing>=0) {
			children.add(state.children.get(routing));
		}
		
		// First we call sendFeebackAsync on children
		for (PredictiveUnitState child : children){
			returns.add(child.predictiveUnitBean.sendFeedbackAsync(feedback,child));
		}
		
		// Then we wait for our own feedback
		state.predictiveUnitBean.doSendFeedback(feedback, state);
		
		//Then we wait for children feedback
		for (Future<Boolean> ret : returns){
			ret.get();
		}
		
		// Finally we store the feedback metrics
		state.predictiveUnitBean.doStoreFeedbackMetrics(feedback,state);
		
		return new AsyncResult<>(true);
	}
	
	protected SeldonMessage transformOutput(SeldonMessage output, PredictiveUnitState state){
		// Transforms the aggregated output into a new message.
		return output;
	}
	
	protected SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state){
		// Transforms the input of a predictive unit into a new message. 
		// The result will become the input of the children, or the output if no children
		return input;
	}
	
	protected SeldonMessage aggregateOutputs(List<SeldonMessage> outputs, PredictiveUnitState state){
		// Aggregates the outputs of all children into a new message. 
		// If there are several outputs, this implementation needs to be overridden.
		
		// TODO: Throw exception if length(outputs) != 1
		return outputs.get(0);
	}
	
	protected int route(SeldonMessage request, PredictiveUnitState state){
		return -1;
	}
	
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		return;
	}
	
	protected void doStoreFeedbackMetrics(Feedback feedback, PredictiveUnitState state){
		return;
	}
	
	private void sanityCheckRouting(Integer branchIndex, PredictiveUnitState state){
		if (branchIndex < -1 | branchIndex >= state.children.size()){
			throw new APIException(
					APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,
					"Invalid branch index. Router that caused the exception: id="+state.name+" name="+state.name);
		}
	}

}
