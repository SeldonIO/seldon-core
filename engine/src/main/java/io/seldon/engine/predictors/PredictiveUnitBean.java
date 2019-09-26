/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;

import org.ojalgo.matrix.PrimitiveMatrix;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;
import org.springframework.stereotype.Component;

import com.google.protobuf.InvalidProtocolBufferException;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Metrics;
import io.micrometer.core.instrument.Tag;
import io.micrometer.core.instrument.Timer;
import io.opentracing.Span;
import io.seldon.engine.exception.APIException;
import io.seldon.engine.metrics.CustomMetricsManager;
import io.seldon.engine.metrics.SeldonRestTemplateExchangeTagsProvider;
import io.seldon.engine.service.InternalPredictionService;
import io.seldon.engine.tracing.TracingProvider;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitMethod;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.Metric;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@Component
public class PredictiveUnitBean extends PredictiveUnitImpl {

	private final static Logger logger = LoggerFactory.getLogger(PredictiveUnitBean.class);
	
	@Autowired
	InternalPredictionService internalPredictionService;
	
	@Autowired
	private SeldonRestTemplateExchangeTagsProvider tagsProvider;
	
	@Autowired
	public PredictorConfigBean predictorConfig;
	
	@Autowired
	private CustomMetricsManager customMetricsManager;
	
	@Autowired
	private TracingProvider tracing;
	
	private PredictiveUnitBeanProxy predictiveUnitBeanProxy;
	
	public PredictiveUnitBean(){}

	public void setProxy(PredictiveUnitBeanProxy proxy)
	{
		this.predictiveUnitBeanProxy = proxy;
	}
	
	public SeldonMessage getOutput(SeldonMessage request, PredictiveUnitState state) throws InterruptedException, ExecutionException, InvalidProtocolBufferException{
		Map<String,Integer> routingDict = new ConcurrentHashMap<String,Integer>();
		Map<String,String> requestPathDict = new ConcurrentHashMap<String,String>();
		Map<String,List<Metric>> metrics = new ConcurrentHashMap<String,List<Metric>>();
		Span activeSpan = null;
		if (tracing != null && tracing.isActive())
			activeSpan = tracing.getTracer().activeSpan();
		SeldonMessage response = predictiveUnitBeanProxy.getOutputAsync(request, state, routingDict,requestPathDict,metrics,activeSpan).get();
		List<Metric> metricList = new ArrayList<>();
		for(List<Metric> mlist: metrics.values())
			metricList.addAll(mlist);
		SeldonMessage.Builder builder = SeldonMessage
			.newBuilder(response)
			.setMeta(Meta
					.newBuilder(response.getMeta()).putAllRouting(routingDict).putAllRequestPath(requestPathDict).addAllMetrics(metricList));
		return builder.build();
	}
	
	private void addMetrics(SeldonMessage msg,PredictiveUnitState state,Map<String,List<Metric>> metrics)
	{
		if (msg.hasMeta())
		{
			addCustomMetrics(msg.getMeta().getMetricsList(),state);
			if (!metrics.containsKey(state.name))
				metrics.putIfAbsent(state.name,new ArrayList<>());
			List<Metric> current = metrics.get(state.name);
			current.addAll(msg.getMeta().getMetricsList());
			metrics.put(state.name,current);
		}
	}
	
	@Async
	public Future<SeldonMessage> getOutputAsync(SeldonMessage input, PredictiveUnitState state, Map<String,Integer> routingDict,Map<String,String> requestPathDict,Map<String,List<Metric>> metrics,Span activeSpan) throws InterruptedException, ExecutionException, InvalidProtocolBufferException{

		String puid = input.getMeta().getPuid();
		
		if (activeSpan != null && tracing != null)
			tracing.getTracer().scopeManager().activate(activeSpan);
		
		// This element to the request path
		requestPathDict.put(state.name, state.image);
		
		// Getting the actual implementation (microservice or hardcoded? )
		PredictiveUnitImpl implementation = predictorConfig.getImplementation(state);
		if (implementation == null){ implementation = this; }
		
		// Compute the transformed Input
		SeldonMessage transformedInput = implementation.transformInput(input, state);
		
		addMetrics(transformedInput,state,metrics);

		// Override the input metadata with the new metadata from transformed input
		transformedInput = mergeMeta(transformedInput, input, puid);

		
		if (state.children.isEmpty()){
			// If this unit has no children, the transformed input becomes the output
			return new AsyncResult<>(transformedInput);
		}
		
		List<PredictiveUnitState> selectedChildren = new ArrayList<PredictiveUnitState>();
		List<Future<SeldonMessage>> deferredChildrenOutputs = new ArrayList<Future<SeldonMessage>>();
		List<SeldonMessage> childrenOutputs = new ArrayList<SeldonMessage>();
		
		// Get the routing. -1 means all children
		SeldonMessage routingMessage = implementation.route(transformedInput, state);
		int routing;
		if (routingMessage != null) {
			routing = getBranchIndex(routingMessage, state);
			sanityCheckRouting(routing, state);
			addMetrics(routingMessage,state,metrics);
		} else {
			routing = -1;
		}
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
		int childIdx = 0;
		for (PredictiveUnitState childState : selectedChildren){
			deferredChildrenOutputs.add(predictiveUnitBeanProxy.getOutputAsync(transformedInput,childState,routingDict,requestPathDict,metrics,activeSpan));
			childIdx++;
		}
		childIdx = 0;
		for (Future<SeldonMessage> deferredOutput : deferredChildrenOutputs){
			SeldonMessage m = deferredOutput.get();
			childrenOutputs.add(m);
			childIdx++;
		}
		
		// Compute the backward transformation of all children outputs
		SeldonMessage aggregatedOutput = implementation.aggregate(childrenOutputs, state);
		addMetrics(aggregatedOutput,state,metrics);
		
		// Merge all the outputs metadata
		aggregatedOutput = mergeMeta(aggregatedOutput, childrenOutputs, puid);
		SeldonMessage transformedOutput = implementation.transformOutput(aggregatedOutput, state);
		addMetrics(transformedOutput,state,metrics);

		// Override the input metadata with the new metadata from transformed input
		transformedOutput = mergeMeta(transformedOutput, aggregatedOutput, puid);

		
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
		PredictiveUnitImpl implementation = predictorConfig.getImplementation(state);
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
		doStoreFeedbackMetrics(feedback,state);
		
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
	
	public SeldonMessage route(SeldonMessage input, PredictiveUnitState state) throws InvalidProtocolBufferException{
		// Returns a branch number in SeldonMessage
		
		if (predictorConfig.hasMethod(PredictiveUnitMethod.ROUTE, state)){
			return internalPredictionService.route(input, state);
		}
		
		// Return default routing
		return null;
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
			PrimitiveMatrix dataArray = PredictorUtils.getOJMatrix(routerReturn.getData());
			branchIndex = dataArray.get(0).intValue();
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
	
	private void addCustomMetrics(List<Metric> metrics, PredictiveUnitState state)
	{
		logger.debug("Add metrics");
		for(Metric metric : metrics)
		{
			Iterable<Tag> tags = tagsProvider.getModelMetrics(state, metric.getTagsMap());
			switch(metric.getType())
			{
			case COUNTER:
				logger.debug("Adding counter {} for {}",metric.getKey(),state.name);
				Counter counter = customMetricsManager.getCounter(tags, metric);
				counter.increment(metric.getValue());
				break;
			case GAUGE:
				logger.debug("Adding gauge {} for {}",metric.getKey(),state.name);		
				customMetricsManager.getGaugeValue(tags, metric).set(metric.getValue());
				break;
			case TIMER:
				logger.debug("Adding timer {} for {}",metric.getKey(),state.name);
				Timer timer = customMetricsManager.getTimer(tags, metric);
				timer.record((long) metric.getValue(), TimeUnit.MILLISECONDS);
				break;
			case UNRECOGNIZED:
				break;
			}
		}
	}		
		
	private void sanityCheckRouting(Integer branchIndex, PredictiveUnitState state){
		if (branchIndex < -1 | branchIndex >= state.children.size()){
			throw new APIException(
					APIException.ApiExceptionType.ENGINE_INVALID_ROUTING,
					"Invalid branch index. Router that caused the exception: id="+state.name+" name="+state.name);
		}
	}

	private SeldonMessage mergeMeta(SeldonMessage latest, List<SeldonMessage> previous, String puid) {
		Meta.Builder metaBuilder = Meta.newBuilder();
		metaBuilder.setPuid(puid);
		for (SeldonMessage originalMessage : previous){
			metaBuilder.putAllTags(originalMessage.getMeta().getTagsMap());
		}
		metaBuilder.putAllTags(latest.getMeta().getTagsMap());
		metaBuilder.clearMetrics();
		return SeldonMessage.newBuilder(latest).setMeta(metaBuilder).build();
	}
	
	private SeldonMessage mergeMeta(SeldonMessage latest, SeldonMessage previous, String puid) {
		Meta.Builder metaBuilder = Meta.newBuilder();
		metaBuilder.setPuid(puid);
		metaBuilder.putAllTags(previous.getMeta().getTagsMap());
		metaBuilder.putAllTags(latest.getMeta().getTagsMap());
		metaBuilder.clearMetrics();
		return SeldonMessage.newBuilder(latest).setMeta(metaBuilder).build();
	}

}
