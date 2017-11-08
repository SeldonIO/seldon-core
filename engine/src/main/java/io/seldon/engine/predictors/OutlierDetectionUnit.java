package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

import io.seldon.protos.PredictionProtos.RequestDef;
import io.seldon.protos.PredictionProtos.ResponseDef;
import io.seldon.protos.PredictionProtos.MetaDef;
import io.seldon.protos.PredictionProtos.OutlierStatus;

public class OutlierDetectionUnit extends PredictiveUnitBean {

	public OutlierDetectionUnit() {
		super();
	}
	
	@Override
	protected ResponseDef backwardPass(List<ResponseDef> inputs, RequestDef request, PredictiveUnitState state){
		
		ResponseDef response = inputs.get(0);
		ResponseDef outlierDetectionResponse = null;
		
		try {
			outlierDetectionResponse = internalPredictionService.getPrediction(request, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		
		Boolean isOutlier = outlierDetectionResponse.getResponse().getTensor().getValues(0) == 1.;
		Double outlierScore = outlierDetectionResponse.getResponse().getTensor().getValues(1);
		
		
		ResponseDef.Builder builder = ResponseDef
	    		.newBuilder(response)
	    		.setMeta(MetaDef
	    				.newBuilder(response.getMeta()).setOutlierStatus(OutlierStatus.newBuilder().setIsOutlier(isOutlier).setScore(outlierScore)));
		
		return builder.build();
	
	}
	
}
