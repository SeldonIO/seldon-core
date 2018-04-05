package io.seldon.wrapper.api.router;

import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.SeldonMessage;

public interface SeldonRouterHandler {
	public SeldonMessage route(SeldonMessage payload);
	public SeldonMessage sendFeedback(Feedback payload);	
}
