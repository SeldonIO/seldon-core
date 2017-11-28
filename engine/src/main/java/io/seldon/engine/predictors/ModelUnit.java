package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
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
import io.seldon.engine.metrics.SeldonRestTemplateExchangeTagsProvider;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Message;

@Component
public class ModelUnit extends PredictiveUnitBean{
	
	@Autowired
	private SeldonRestTemplateExchangeTagsProvider tagsProvider;
	
	public ModelUnit() {
		super();
	}
	
	private Message doPredict(Message request, PredictiveUnitState state)
	{
		Message ret = null;
		try {
			ret = internalPredictionService.getPrediction(request, state);
//			ret = Message.newBuilder(ret).setMeta(Meta.newBuilder(ret.getMeta()).addModel(state.id)).build();
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return ret;
	}
	
	@Override
	protected void doStoreFeedbackMetrics(Feedback feedback, PredictiveUnitState state){

		Counter.builder("seldon_api_model_feedback_reward").tags(tagsProvider.getModelMetrics(state)).register(Metrics.globalRegistry).increment(feedback.getReward());
		Counter.builder("seldon_api_model_feedback").tags(tagsProvider.getModelMetrics(state)).register(Metrics.globalRegistry).increment();
	}

	@Override
	@Async
	protected Future<Message> predict(Message request, PredictiveUnitState state, Map<String,Integer> routingDict) throws InterruptedException, ExecutionException{
		return new AsyncResult<>(doPredict(request,state));
	}

	@Override
	protected List<PredictiveUnitState> forwardPass(Message request, PredictiveUnitState state, Map<String,Integer> routingDict){
		return new ArrayList<PredictiveUnitState>(Arrays.asList(state));
	}
	
	
}
