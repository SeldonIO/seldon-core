package io.seldon.engine.predictors;

public class PredictorRequestJSON extends PredictorRequest{
	public String[] features;
	public double[][] values;
	
	public PredictorRequestJSON(){}
	
}
