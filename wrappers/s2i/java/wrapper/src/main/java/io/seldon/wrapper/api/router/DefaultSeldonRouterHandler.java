package io.seldon.wrapper.api.router;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

@Component
public class DefaultSeldonRouterHandler implements SeldonRouterHandler {

	@Override
	public SeldonMessage route(SeldonMessage payload) {
		// TODO Auto-generated method stub
		return null;
	}

	@Override
	public SeldonMessage sendFeedback(Feedback payload) {
		// TODO Auto-generated method stub
		return null;
	}

	
}
