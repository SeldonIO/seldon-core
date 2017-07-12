package io.seldon.engine.predictors;

import java.util.Arrays;
import java.util.List;
import java.util.Map;

import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;

@Component
public class AverageCombinerUnit extends CombinerUnit{
	
	public AverageCombinerUnit() {}

	@Override
	public PredictorReturn backwardPass(List<PredictorData> inputs, PredictiveUnitState state){
		
		Integer batchLength = 0;
		Integer valuesLength = 0;
		Integer inputsLength = inputs.size();
		Boolean initialised = false;
		Double[][] averages = null;
		String[] names = null;
		
		for (PredictorData predData : inputs){
			PredictorReturn predRet = (PredictorReturn) predData;
			if (!initialised){
				batchLength = predRet.values.length;
				valuesLength = predRet.values[0].length;
				averages = new Double[batchLength][valuesLength];
				for (int i =0; i < batchLength; i++){
					Arrays.fill(averages[i], 0.);
				}	
				names = predRet.names;
				initialised = true;
			}
			if (predRet.values.length!=batchLength){
				// TODO: Maybe we should also check that the names are always the same
				throw new APIException(APIException.ApiExceptionType.APIFE_INVALID_COMBINER_RESPONSE,String.format("Found %d Expected %d", predRet.values[0].length,valuesLength));				
			}
			if (predRet.values[0].length!=valuesLength){
				throw new APIException(APIException.ApiExceptionType.APIFE_INVALID_COMBINER_RESPONSE,String.format("Found %d Expected %d", predRet.values[0].length,valuesLength));
			}
			for (int i = 0; i < batchLength; ++i) {
				for (int j = 0; j < valuesLength; j++){
					averages[i][j] += predRet.values[i][j];
				}
			}
		}
		
		for (int i = 0; i < batchLength; ++i) {
			for (int j = 0; j < valuesLength; j++){
				averages[i][j] /= inputsLength;
			}
		}
		
		return new PredictorReturn(names,averages);
	}

}
