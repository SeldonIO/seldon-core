package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.List;
import java.util.Random;

import org.springframework.stereotype.Component;

import io.seldon.engine.exception.APIException;
import io.seldon.protos.PredictionProtos.Feedback;
import io.seldon.protos.PredictionProtos.Message;


@Component
public class RandomABTestUnit extends RouterUnit {
	
	Random rand = new Random(1337);

	@Override
	protected Integer forwardPass(Message request, PredictiveUnitState state){
		
		@SuppressWarnings("unchecked")
		PredictiveUnitParameter<Float> parameter = (PredictiveUnitParameter<Float>) state.parameters.get("ratioA");
		
		if (parameter == null){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ABTEST,"Parameter 'ratioA' is missing.");
		}
		
		Float ratioA = parameter.value;
		Float comparator = rand.nextFloat();
		
		if (state.children.size() != 2){
			throw new APIException(APIException.ApiExceptionType.ENGINE_INVALID_ABTEST,String.format("AB test has %d children ",state.children.size()));
		}
		
		//FIXME Possible bug : keySet is not ordered as per the definition of the AB test
		if (comparator<=ratioA){
			// We select model A
			return 0;
		}
		else{
			return 1;
		}
	} 
	
	@Override
	protected void doSendFeedback(Feedback feedback, PredictiveUnitState state){
		return;
	}
}
