package io.seldon.engine.predictors;

import java.io.IOException;

import org.springframework.stereotype.Component;

import com.google.protobuf.Value;

import io.seldon.protos.PredictionProtos.SeldonMessage;
import io.seldon.protos.PredictionProtos.Meta;

@Component
public class OutlierDetectionUnit extends PredictiveUnitBean {

	public OutlierDetectionUnit() {
		super();
	}
	
	@Override
	protected SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state){
		
		SeldonMessage outlierDetectionResponse = null;
		
		try {
			outlierDetectionResponse = internalPredictionService.scoreOutlier(input, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		
		Double outlierScore = outlierDetectionResponse.getData().getTensor().getValues(0);
		
		Value outlierScoreValue = Value.newBuilder().setNumberValue(outlierScore).build();
		
		SeldonMessage.Builder builder = SeldonMessage
	    		.newBuilder(input)
	    		.setMeta(Meta
	    				.newBuilder(input.getMeta()).putTags("outlierScore", outlierScoreValue)
	    				);
		
		return builder.build();
	}
	
}
