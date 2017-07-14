package io.seldon.engine.predictors;

public class PredictorRequest  {
	
	public String request;
	
	
	public PredictorRequest(){}
	
	public PredictorRequest(String request){
		this.request = request;
	}

//	public PredictorRequest(String jsonRaw){
//		request = jsonRaw;
//	}

//	public PredictorRequest(String request) {
//		super();
//		this.request = request;
//	}
//	
	
}
