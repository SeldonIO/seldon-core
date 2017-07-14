package io.seldon.engine.service;

import io.seldon.engine.predictors.PredictorReturn;

public class PredictionServiceReturn {
	
	public PredictionServiceReturnStatus status;
	public PredictionServiceReturnMeta meta;
	public PredictorReturn response;
	
	public PredictionServiceReturn(PredictionServiceReturnMeta meta, PredictorReturn response){
		this.status = new PredictionServiceReturnStatus();
		this.meta = meta;
		this.response = response;
	}
	
	public PredictionServiceReturn(){
		this.meta = new PredictionServiceReturnMeta("");
		this.status = new PredictionServiceReturnStatus();
		this.response = new PredictorReturn(new String[0], new Double[0][0]);
	}

}
