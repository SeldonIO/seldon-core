package io.seldon.engine.service;

import io.seldon.engine.predictors.PredictorReturn;

public class PredictionServiceReturn {
	
	public PredictionServiceReturnMeta meta;
	public PredictorReturn predictorReturn;
	
	public PredictionServiceReturn(PredictionServiceReturnMeta meta, PredictorReturn predictorReturn){
		this.meta = meta;
		this.predictorReturn = predictorReturn;
	}
	
	public PredictionServiceReturn(){
		this.meta = new PredictionServiceReturnMeta();
		this.predictorReturn = new PredictorReturn(new String[0], new Double[0][0]);
	}

}
