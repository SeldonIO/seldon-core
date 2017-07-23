package io.seldon.engine.predictors;

import java.util.Arrays;
import java.util.List;

import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.DefaultDataDef;
import io.seldon.protos.PredictionProtos.DefaultDataValues;
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
		Double[][] averages = null;
		
		PredictionResponseDef.Builder respBuilder = PredictionResponseDef.newBuilder();
		DefaultDataDef.Builder dataBuilder = DefaultDataDef.newBuilder();
		int modelIdx = 0;
		for (PredictionResponseDef predRet : inputs){
			if (!initialised){
				batchLength = predRet.getResponse().getValuesCount();
				valuesLength = predRet.getResponse().getValues(0).getValueCount();
				averages = new Double[batchLength][valuesLength];
				for (int i =0; i < batchLength; i++){
					Arrays.fill(averages[i], 0.);
				}	
				respBuilder.setMeta(predRet.getMeta()).setStatus(predRet.getStatus());
				dataBuilder.addAllNames(predRet.getResponse().getNamesList());
				initialised = true;
			}
			if (predRet.getResponse().getValuesCount()!=batchLength){
				// TODO: Maybe we should also check that the names are always the same
				throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,String.format("Found batch size %d Expected %d for model %d", predRet.getResponse().getValuesCount(),valuesLength,modelIdx));				
			}
			if (predRet.getResponse().getValues(0).getValueCount()!=valuesLength){
				throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_COMBINER_RESPONSE,String.format("Found value size %d Expected %d for model %d", predRet.getResponse().getValues(0).getValueCount(),valuesLength,modelIdx));
			}
			for (int i = 0; i < batchLength; ++i) {
				for (int j = 0; j < valuesLength; j++){
					averages[i][j] += predRet.getResponse().getValues(i).getValue(j);
				}
			}
			modelIdx++;
		}
		
		for (int i = 0; i < batchLength; ++i) {
			for (int j = 0; j < valuesLength; j++){
				averages[i][j] /= inputsLength;
			}
		}
	
		for (int i = 0; i < batchLength; ++i) {
			dataBuilder.addValues(DefaultDataValues.newBuilder().addAllValue(Arrays.asList(averages[i])).build());
		}
		respBuilder.setResponse(dataBuilder.build());
		
		return respBuilder.build();
	}

}
