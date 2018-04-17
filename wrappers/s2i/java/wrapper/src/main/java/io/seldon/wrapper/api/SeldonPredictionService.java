package io.seldon.wrapper.api;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.SeldonMessageList;

public interface SeldonPredictionService {
	default public SeldonMessage predict(SeldonMessage request) {
		return null;
	}
	default public SeldonMessage route(SeldonMessage request) {
		return null;
	}
	default public SeldonMessage sendFeedback(Feedback request) {
		return null;
	}
	default public SeldonMessage transformInput(SeldonMessage request) {
		return null;
	}
	default public SeldonMessage transformOutput(SeldonMessage request) {
		return null;
	}
	default public SeldonMessage aggregate(SeldonMessageList request) {
		return null;
	}
}
