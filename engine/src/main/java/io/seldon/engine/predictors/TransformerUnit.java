package io.seldon.engine.predictors;

import java.io.IOException;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.Message;

@Component
public class TransformerUnit extends PredictiveUnitBean {

	public TransformerUnit(){
		super();
	}
	
	@Override
	protected Message transformInput(Message input, PredictiveUnitState state){
		Message transformedInput = null;
		try {
			transformedInput = internalPredictionService.transformInput(input, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return transformedInput;
	}
	
	@Override
	protected Message transformOutput(Message output, PredictiveUnitState state){
		Message transformedOutput = null;
		try {
			transformedOutput = internalPredictionService.transformOutput(output, state);
		} catch(IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return transformedOutput;
	}
}
