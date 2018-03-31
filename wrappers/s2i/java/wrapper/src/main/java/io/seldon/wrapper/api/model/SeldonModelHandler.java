package io.seldon.wrapper.api.model;

import io.seldon.protos.PredictionProtos.SeldonMessage;

public interface SeldonModelHandler {

	public SeldonMessage predict(SeldonMessage payload);
	
}
