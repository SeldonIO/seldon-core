package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;

import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.PredictionRequestDef;


@Component
public class RandomABTestUnit extends RouterUnit {
	
	Random rand = new Random(1337);

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictionRequestDef request, PredictiveUnitState state){
		
		@SuppressWarnings("unchecked")
		PredictiveUnitParameter<Float> parameter = (PredictiveUnitParameter<Float>) state.parameters.get("ratioA");
		
		Float ratioA = parameter.value;
		Float comparator = rand.nextFloat();
		
		if (state.children.size() != 2){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ABTEST,String.format("test has %d children ",state.children.size()));
		}
		
		PredictiveUnitState selectedChild;
		
		//FIXME Possible bug : keySet is not ordered as per the definition of the AB test
		if (comparator<=ratioA){
			// We select model A
			selectedChild = state.children.get(state.children.keySet().toArray()[0]);
		}
		else{
			selectedChild = state.children.get(state.children.keySet().toArray()[1]);
		}
		
		List<PredictiveUnitState> ret = new ArrayList<>();
		ret.add(selectedChild);
		return ret;
	} 
}
