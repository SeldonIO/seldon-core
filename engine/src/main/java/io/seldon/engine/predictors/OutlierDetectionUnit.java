package io.seldon.engine.predictors;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

import io.seldon.protos.PredictionProtos.Request;
import io.seldon.protos.PredictionProtos.Response;
import io.seldon.protos.PredictionProtos.Meta;
import io.seldon.protos.PredictionProtos.OutlierStatus;

public class OutlierDetectionUnit extends PredictiveUnitBean {

	public OutlierDetectionUnit() {
		super();
	}
	
	@Override
	protected Response backwardPass(List<Response> inputs, Request request, PredictiveUnitState state){
		
		Response response = inputs.get(0);
		Response outlierDetectionResponse = null;
		
		try {
			outlierDetectionResponse = internalPredictionService.getPrediction(request, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		
		Boolean isOutlier = outlierDetectionResponse.getData().getTensor().getValues(0) == 1.;
		Double outlierScore = outlierDetectionResponse.getData().getTensor().getValues(1);
		
		
		Response.Builder builder = Response
	    		.newBuilder(response)
	    		.setMeta(Meta
	    				.newBuilder(response.getMeta()).setOutlierStatus(OutlierStatus.newBuilder().setIsOutlier(isOutlier).setScore(outlierScore)));
		
		return builder.build();
	
	}
	
}
