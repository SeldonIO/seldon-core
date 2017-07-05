package io.seldon.apife.predictors;

public class PredictorRequestJSON extends PredictorRequest{
	public String data;
	
	public PredictorRequestJSON(){}
	
	public PredictorRequestJSON(String data){
		this.data = data;
	}
}
