package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;

import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;


@Component
public class RandomABTestUnit extends RouterUnit {
	
	Random rand = new Random(1337);

	@Override
	protected List<PredictiveUnitState> forwardPass(PredictorData request, PredictiveUnitState state){
		
		@SuppressWarnings("unchecked")
		PredictiveUnitParameter<Float> parameter = (PredictiveUnitParameter<Float>) state.parameters.get("ratioA");
		
		Float ratioA = parameter.value;
		Float comparator = rand.nextFloat();
		
		if (state.children.size() != 2){
			throw new APIException(APIException.ABTEST_ERROR);
		}
		
		PredictiveUnitState selectedChild;
		
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
