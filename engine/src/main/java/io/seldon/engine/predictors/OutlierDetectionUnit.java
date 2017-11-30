package io.seldon.engine.predictors;

import java.io.IOException;
import io.seldon.protos.PredictionProtos.Message;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.OutlierStatus;

public class OutlierDetectionUnit extends PredictiveUnitBean {

	public OutlierDetectionUnit() {
		super();
	}
	
	@Override
	protected Message transformInput(Message input, PredictiveUnitState state){
		
		Message outlierDetectionResponse = null;
		
		try {
			outlierDetectionResponse = internalPredictionService.getPrediction(input, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		
		Boolean isOutlier = outlierDetectionResponse.getData().getTensor().getValues(0) == 1.;
		Double outlierScore = outlierDetectionResponse.getData().getTensor().getValues(1);
		
		Message.Builder builder = Message
	    		.newBuilder(input)
	    		.setMeta(Meta
	    				.newBuilder(input.getMeta()).setOutlierStatus(OutlierStatus.newBuilder().setIsOutlier(isOutlier).setScore(outlierScore)));
		
		return builder.build();
	}
	
}
