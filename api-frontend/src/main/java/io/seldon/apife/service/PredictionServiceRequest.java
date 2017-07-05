package io.seldon.apife.service;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

import io.seldon.apife.predictors.PredictorRequest;
import io.seldon.apife.serializers.PredictionServiceRequestDeserializer;

@JsonDeserialize(using = PredictionServiceRequestDeserializer.class)
public class PredictionServiceRequest {
	
	@JsonProperty("meta")
	public PredictionServiceRequestMeta meta;
	
	@JsonProperty("request")
	public PredictorRequest request;

	public PredictionServiceRequestMeta getMeta(){
		return meta;
	}
	
	public void setMeta(PredictionServiceRequestMeta meta){
		this.meta = meta;
	}
	
	public PredictorRequest getRequest(){
		return request;
	}
	
	public void setRequest(PredictorRequest request){
		this.request = request;
	}
	
	public PredictionServiceRequest(PredictionServiceRequestMeta meta, PredictorRequest request){
		this.meta = meta;
		this.request = request;
	}
	
}
