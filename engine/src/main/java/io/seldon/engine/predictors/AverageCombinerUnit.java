package io.seldon.engine.predictors;

import java.util.Arrays;
import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.PredictionResponseDef;

@Component
public class AverageCombinerUnit extends CombinerUnit{
	
	public AverageCombinerUnit() {}

	@Override
	public PredictionResponseDef backwardPass(List<PredictionResponseDef> inputs, PredictiveUnitState state){
		
		Integer batchLength = 0;
		Integer valuesLength = 0;
		Integer inputsLength = inputs.size();
		Boolean initialised = false;
		Double[] averages = null;
		
		PredictionResponseDef.Builder respBuilder = PredictionResponseDef.newBuilder();
		DefaultDataDef.Builder dataBuilder = DefaultDataDef.newBuilder();
		int modelIdx = 0;
		for (PredictionResponseDef predRet : inputs){
			if (!initialised){
				if (predRet.getResponse().getShapeCount() == 2)
				{
					batchLength = predRet.getResponse().getShape(0);
					valuesLength = predRet.getResponse().getShape(1);
				}
				else
				{
					batchLength = 1;
					valuesLength = predRet.getResponse().getValuesCount();
				}
				
				averages = new Double[batchLength*valuesLength];
				Arrays.fill(averages, 0.);
				
				respBuilder.setMeta(predRet.getMeta()).setStatus(predRet.getStatus());
				dataBuilder.addAllKeys(predRet.getResponse().getKeysList());
				initialised = true;
			}
			else
			{
				if (predRet.getResponse().getValuesCount() != (averages.length))
					throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE, String.format("Expected values of length %d but found %d",averages.length,predRet.getResponse().getValuesCount()));
			}
			for (int i = 0; i < batchLength; ++i) {
				for (int j = 0; j < valuesLength; j++){
					averages[(i*valuesLength)+j] += predRet.getResponse().getValues((i*valuesLength)+j);
				}
			}
			modelIdx++;
		}
		
		for (int i = 0; i < batchLength; ++i) {
			for (int j = 0; j < valuesLength; j++){
				averages[(i*valuesLength)+j] /= inputsLength;
			}
		}
	
		if (averages != null)
			dataBuilder.addAllValues(Arrays.asList(averages)).build();
		respBuilder.setResponse(dataBuilder.build());
		
		return respBuilder.build();
	}

}
