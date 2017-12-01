package io.seldon.engine.predictors;

import java.io.IOException;

import org.springframework.stereotype.Component;

import io.seldon.protos.PredictionProtos.SeldonMessage;

@Component
public class TransformerUnit extends PredictiveUnitBean {

	public TransformerUnit(){
		super();
	}
	
	@Override
	protected SeldonMessage transformInput(SeldonMessage input, PredictiveUnitState state){
		SeldonMessage transformedInput = null;
		try {
			transformedInput = internalPredictionService.transformInput(input, state);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return transformedInput;
	}
	
	@Override
	protected SeldonMessage transformOutput(SeldonMessage output, PredictiveUnitState state){
		SeldonMessage transformedOutput = null;
		try {
			transformedOutput = internalPredictionService.transformOutput(output, state);
		} catch(IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return transformedOutput;
	}
}
