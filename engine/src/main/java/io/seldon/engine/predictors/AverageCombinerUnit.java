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
		
		Integer valuesLength = 0;
		Integer inputsLength = inputs.size();
		Boolean initialised = false;
		Double[] averages = null;
		String[] names = null;
		
		for (PredictorData predData : inputs){
			PredictorReturn predRet = (PredictorReturn) predData;
			if (!initialised){
				valuesLength = predRet.values.length;
				averages = new Double[valuesLength];
				Arrays.fill(averages, 0.);
				names = predRet.names;
				initialised = true;
			}
			if (predRet.values.length!=valuesLength){
				// TODO: Maybe we should also check that the names are always the same
				throw new APIException(APIException.COMBINER_ERROR);
			}
			for (int i = 0; i < valuesLength; ++i) {
			    averages[i] += predRet.values[i];
			}
		}
		
		for (int i = 0; i < valuesLength; ++i) {
		    averages[i] /= inputsLength;
		}
		
		return new PredictorReturn(names,averages);
	}

}
