package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

import io.seldon.protos.PredictionProtos.PredictionRequestDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;
import io.seldon.protos.PredictionProtos.PredictionResponseMetaDef;
import io.seldon.protos.PredictionProtos.OutlierStatus;

public class OutlierDetectionUnit extends PredictiveUnitBean {

	public OutlierDetectionUnit() {
		super();
	}
	
	@Override
	protected PredictionResponseDef backwardPass(List<PredictionResponseDef> inputs, PredictionRequestDef request, PredictiveUnitState state){
		
		PredictionResponseDef response = inputs.get(0);
		PredictionResponseDef outlierDetectionResponse = null;
		
		try {
			outlierDetectionResponse = internalPredictionService.getPrediction(request, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		
		Boolean isOutlier = outlierDetectionResponse.getResponse().getTensor().getValues(0) == 1.;
		Double outlierScore = outlierDetectionResponse.getResponse().getTensor().getValues(1);
		
		
		PredictionResponseDef.Builder builder = PredictionResponseDef
	    		.newBuilder(response)
	    		.setMeta(PredictionResponseMetaDef
	    				.newBuilder(response.getMeta()).setOutlierStatus(OutlierStatus.newBuilder().setIsOutlier(isOutlier).setScore(outlierScore)));
		
		return builder.build();
	
	}
	
}
