package io.seldon.engine.predictors;

import io.seldon.protos.DeploymentProtos.Parameter;
import io.seldon.engine.predictors.PredictiveUnitParameterInterface;

public class PredictiveUnitParameter<T> extends PredictiveUnitParameterInterface{
	public T value;
	
	public PredictiveUnitParameter(T value){
		this.value = value;
	}
	
	public static PredictiveUnitParameterInterface fromParameter(Parameter Parameter){
		switch (Parameter.getType()){
			case DOUBLE:
				Double valueDouble = Double.parseDouble(Parameter.getValue());
				return new PredictiveUnitParameter<Double>(valueDouble);
			case FLOAT:
				Float valueFloat = Float.parseFloat(Parameter.getValue());
				return new PredictiveUnitParameter<Float>(valueFloat);
			case INT:
				Integer valueInt = Integer.parseInt(Parameter.getValue());
				return new PredictiveUnitParameter<Integer>(valueInt);
			case STRING:
				return new PredictiveUnitParameter<String>(Parameter.getValue());			
			default:
				return new PredictiveUnitParameter<String>(Parameter.getValue());
		}
	}
}
