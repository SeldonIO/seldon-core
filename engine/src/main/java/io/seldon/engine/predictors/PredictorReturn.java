package io.seldon.engine.predictors;

public class PredictorReturn {
	public String[] names;
	public Double[][] values;
	
	public PredictorReturn(){}
	
	public PredictorReturn(String[] names, Double[][] values){
		this.names = names;
		this.values = values;
	}
}
