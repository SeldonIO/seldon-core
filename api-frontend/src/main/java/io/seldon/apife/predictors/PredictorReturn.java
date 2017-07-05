package io.seldon.apife.predictors;

public class PredictorReturn extends PredictorData{
	public String[] names;
	public Double[][] values;
	
	public PredictorReturn(){}
	
	public PredictorReturn(String[] names, Double[][] values){
		this.names = names;
		this.values = values;
	}
}
