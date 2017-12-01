package io.seldon.engine.predictors;

import java.io.IOException;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Metrics;
import io.seldon.engine.metrics.SeldonRestTemplateExchangeTagsProvider;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@Component
public class ModelUnit extends PredictiveUnitBean{
	
	@Autowired
	private SeldonRestTemplateExchangeTagsProvider tagsProvider;
	
	public ModelUnit() {
		super();
	}
	
	@Override
	protected void doStoreFeedbackMetrics(Feedback feedback, PredictiveUnitState state){

		Counter.builder("seldon_api_model_feedback_reward").tags(tagsProvider.getModelMetrics(state)).register(Metrics.globalRegistry).increment(feedback.getReward());
		Counter.builder("seldon_api_model_feedback").tags(tagsProvider.getModelMetrics(state)).register(Metrics.globalRegistry).increment();
	}
	
	@Override
	protected SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state){
		SeldonMessage output = null;
		try {
			output = internalPredictionService.predict(input, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return output;
	}
	
}
