package io.seldon.apife.predictors;

public class PredictorRequest extends PredictorData {
	
	public String request;
	public Boolean isDefault;
	
	public PredictorRequest(){}
	
	public PredictorRequest(String request, Boolean isDefault){
		this.request = request;
		this.isDefault = isDefault;
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
