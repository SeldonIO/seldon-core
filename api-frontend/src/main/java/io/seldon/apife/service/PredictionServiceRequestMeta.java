package io.seldon.apife.service;

import com.fasterxml.jackson.annotation.JsonProperty;

public class PredictionServiceRequestMeta {
	
	@JsonProperty("deployment")
	public String deployment;
	
	public String getDeployment(){
		return deployment;
	}
	
	public void setDeployment(String deployment){
		this.deployment = deployment;
	}

//	public PredictionServiceRequestMeta(String deployment) {
//		super();
//		this.deployment = deployment;
//	}

}
