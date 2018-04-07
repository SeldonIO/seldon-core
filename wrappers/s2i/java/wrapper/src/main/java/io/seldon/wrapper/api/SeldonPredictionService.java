package io.seldon.wrapper.api;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

public interface SeldonPredictionService {
	default public SeldonMessage predict(SeldonMessage request) {
		return null;
	}
	default public SeldonMessage route(SeldonMessage payload) {
		return null;
	}
	default public SeldonMessage sendFeedback(Feedback payload) {
		return null;
	}
	
	
}
