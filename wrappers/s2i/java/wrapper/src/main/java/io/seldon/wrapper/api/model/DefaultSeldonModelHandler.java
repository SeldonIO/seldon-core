package io.seldon.wrapper.api.model;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.SeldonMessage;

@Component
public class DefaultSeldonModelHandler implements SeldonModelHandler {

	@Override
	public SeldonMessage predict(SeldonMessage payload) {
		// TODO Auto-generated method stub
		return null;
	}

}
