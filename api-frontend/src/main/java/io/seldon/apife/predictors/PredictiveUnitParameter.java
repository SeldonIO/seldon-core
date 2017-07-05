package io.seldon.apife.predictors;

import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;

public class PredictiveUnitParameter<T> extends PredictiveUnitParameterInterface{
	public T value;
	
	public PredictiveUnitParameter(T value){
		this.value = value;
	}
	
	public static PredictiveUnitParameterInterface fromParamDef(ParamDef paramDef){
		switch (paramDef.getType()){
			case DOUBLE:
				Double valueDouble = Double.parseDouble(paramDef.getValue());
				return new PredictiveUnitParameter<Double>(valueDouble);
			case FLOAT:
				Float valueFloat = Float.parseFloat(paramDef.getValue());
				return new PredictiveUnitParameter<Float>(valueFloat);
			case INT:
				Integer valueInt = Integer.parseInt(paramDef.getValue());
				return new PredictiveUnitParameter<Integer>(valueInt);
			case STRING:
				return new PredictiveUnitParameter<String>(paramDef.getValue());			
			default:
				return new PredictiveUnitParameter<String>(paramDef.getValue());
		}
	}
}
