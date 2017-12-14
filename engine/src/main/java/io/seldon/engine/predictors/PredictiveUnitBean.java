package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.nd4j.linalg.api.ndarray.INDArray;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import com.google.protobuf.InvalidProtocolBufferException;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Metrics;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.metrics.SeldonRestTemplateExchangeTagsProvider;
import io.seldon.engine.service.InternalPredictionService;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitMethod;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Meta;

@Component
public class PredictiveUnitBean {

	@Autowired
	InternalPredictionService internalPredictionService;
	
	@Autowired
	private SeldonRestTemplateExchangeTagsProvider tagsProvider;
	
	
	@Autowired
	PredictorConfigBean predictorConfig;
	
	public PredictiveUnitBean() {}
	
//	public PredictiveUnitBean(InternalPredictionService internalPredictionService){
//		this.internalPredictionService = internalPredictionService;
//	}
	
	public SeldonMessage getOutput(SeldonMessage request, PredictiveUnitState state) throws InterruptedException, ExecutionException, InvalidProtocolBufferException{
		Map<String,Integer> routingDict = new HashMap<String,Integer>();
		SeldonMessage response = getOutputAsync(request, state, routingDict).get();
		SeldonMessage.Builder builder = SeldonMessage
	    		.newBuilder(response)
	    		.setMeta(Meta
	    				.newBuilder(response.getMeta()).putAllRouting(routingDict));
		return builder.build();
	}
	
	@Async
	private Future<SeldonMessage> getOutputAsync(SeldonMessage input, PredictiveUnitState state, Map<String,Integer> routingDict) throws InterruptedException, ExecutionException, InvalidProtocolBufferException{
		
		// Getting the actual implementation (microservice or hardcoded? )
		PredictiveUnitBean implementation = predictorConfig.getImplementation(state);
		if (implementation == null){ implementation = this; }
		
		// Compute the transformed Input
		SeldonMessage transformedInput = implementation.transformInput(input, state);
		
		// Preserve the original metadata
		transformedInput = mergeMeta(transformedInput,input.getMeta());
		
		if (state.children.isEmpty()){
			// If this unit has no children, the transformed input becomes the output
			return new AsyncResult<>(transformedInput);
		}
		
		List<PredictiveUnitState> selectedChildren = new ArrayList<PredictiveUnitState>();
		List<Future<SeldonMessage>> deferredChildrenOutputs = new ArrayList<Future<SeldonMessage>>();
		List<SeldonMessage> childrenOutputs = new ArrayList<SeldonMessage>();
		
		// Get the routing. -1 means all children
		int routing = implementation.route(transformedInput, state);
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
			deferredChildrenOutputs.add(getOutputAsync(transformedInput,childState,routingDict));
		}
		for (Future<SeldonMessage> deferredOutput : deferredChildrenOutputs){
			childrenOutputs.add(deferredOutput.get());
		}
		
		// Compute the backward transformation of all children outputs
		SeldonMessage aggregatedOutput = implementation.aggregate(childrenOutputs, state);
		// Merge all the outputs metadata
		aggregatedOutput = mergeMeta(aggregatedOutput,childrenOutputs);
		SeldonMessage transformedOutput = implementation.transformOutput(aggregatedOutput, state);
		// Preserve metadata
		transformedOutput = mergeMeta(transformedOutput,aggregatedOutput.getMeta());
		
		return new AsyncResult<>(transformedOutput);
		
	}
	
	public void sendFeedback(Feedback feedback, PredictiveUnitState state) throws InterruptedException, ExecutionException, InvalidProtocolBufferException{
		sendFeedbackAsync(feedback,state).get();
	}
	
	@Async
	private Future<Boolean> sendFeedbackAsync(Feedback feedback, PredictiveUnitState state) throws InterruptedException, ExecutionException, InvalidProtocolBufferException{
		System.out.println("NODE " + state.name + ": entered feedback");
		List<PredictiveUnitState> children = new ArrayList<PredictiveUnitState>();
		List<Future<Boolean>> returns = new ArrayList<Future<Boolean>>();
		
		// Getting the actual implementation (microservice or hardcoded? )
		PredictiveUnitBean implementation = predictorConfig.getImplementation(state);
		if (implementation == null){ implementation = this; }
				
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
			returns.add(sendFeedbackAsync(feedback,child));
		}
		
		// Then we wait for our own feedback
		implementation.doSendFeedback(feedback, state);
		
		//Then we wait for children feedback
		for (Future<Boolean> ret : returns){
			ret.get();
		}
		
		// Finally we store the feedback metrics
		implementation.doStoreFeedbackMetrics(feedback,state);
		
		return new AsyncResult<>(true);
	}
	
	// -----------------------------
	// The "predictive unit methods"
	// -----------------------------
	
	public SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state) throws InvalidProtocolBufferException{
		// Transforms the input of a predictive unit into a new message. 
		// The result will become the input of the children, or the output if no children
		if (predictorConfig.hasMethod(PredictiveUnitMethod.TRANSFORM_INPUT, state)) {
			return internalPredictionService.transformInput(input, state);
		}
		return input;
	}
	
	public SeldonMessage transformOutput(SeldonMessage output, PredictiveUnitState state) throws InvalidProtocolBufferException{
		// Transforms the aggregated output into a new message.
		if (predictorConfig.hasMethod(PredictiveUnitMethod.TRANSFORM_OUTPUT, state)) {
			return internalPredictionService.transformOutput(output, state);
		}
		return output;
	}
	
	public SeldonMessage aggregate(List<SeldonMessage> outputs, PredictiveUnitState state) throws InvalidProtocolBufferException{
		// Aggregates the outputs of all children into a new message. 
		// If there are several outputs, this implementation needs to be overridden.
		
		
		if (predictorConfig.hasMethod(PredictiveUnitMethod.AGGREGATE, state)) {
			return internalPredictionService.aggregate(outputs,state);
		}
		
		// TODO: Throw exception if length(outputs) != 1
		return outputs.get(0);
	}
	
	public int route(SeldonMessage input, PredictiveUnitState state) throws InvalidProtocolBufferException{
		// Returns a branch number
		
		if (predictorConfig.hasMethod(PredictiveUnitMethod.ROUTE, state)){
			SeldonMessage routerReturn = internalPredictionService.route(input, state);
			return getBranchIndex(routerReturn, state);
		}
		return -1;
	}
	
	public void doSendFeedback(Feedback feedback, PredictiveUnitState state) throws InvalidProtocolBufferException{
		// Sends feedback to the microservice 
		
		if (predictorConfig.hasMethod(PredictiveUnitMethod.SEND_FEEDBACK, state)){
			internalPredictionService.sendFeedback(feedback, state);
		}
		return;
	}
	
	// -------------------------------------------
	//
	// -------------------------------------------
	
	private int getBranchIndex(SeldonMessage routerReturn, PredictiveUnitState state){
		int branchIndex = 0;
		try {
			INDArray dataArray = PredictorUtils.getINDArray(routerReturn.getData());
			branchIndex = (int) dataArray.getInt(0);
		}
		catch (IndexOutOfBoundsException e) {
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,"Router that caused the exception: id="+state.name+" name="+state.name);
		}
		return branchIndex;
	}
	
	protected void doStoreFeedbackMetrics(Feedback feedback, PredictiveUnitState state){
		Counter.builder("seldon_api_model_feedback_reward").tags(tagsProvider.getModelMetrics(state)).register(Metrics.globalRegistry).increment(feedback.getReward());
		Counter.builder("seldon_api_model_feedback").tags(tagsProvider.getModelMetrics(state)).register(Metrics.globalRegistry).increment();
	}
	
	private void sanityCheckRouting(Integer branchIndex, PredictiveUnitState state){
		if (branchIndex < -1 | branchIndex >= state.children.size()){
			throw new APIException(
					APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,
					"Invalid branch index. Router that caused the exception: id="+state.name+" name="+state.name);
		}
	}

	private SeldonMessage mergeMeta(SeldonMessage message, List<SeldonMessage> messages) {
		Meta.Builder metaBuilder = Meta.newBuilder(message.getMeta());
		for (SeldonMessage originalMessage : messages){
			metaBuilder.putAllTags(originalMessage.getMeta().getTagsMap());
		}
		return SeldonMessage.newBuilder(message).setMeta(metaBuilder).build();
	}
	
	private SeldonMessage mergeMeta(SeldonMessage message, Meta meta) {
		Meta.Builder metaBuilder = Meta.newBuilder(message.getMeta());
		metaBuilder.putAllTags(meta.getTagsMap());
		return SeldonMessage.newBuilder(message).setMeta(metaBuilder).build();
	}

}
