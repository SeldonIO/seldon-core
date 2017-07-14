package io.seldon.engine.service;

import com.fasterxml.jackson.annotation.JsonProperty;

public class PredictionServiceRequestMeta {
	
	@JsonProperty("puid")
	public String puid;

	public String getPuid() {
		return puid;
	}

	public void setPuid(String puid) {
		this.puid = puid;
	}
	

//	public PredictionServiceRequestMeta(String deployment) {
//		super();
//		this.deployment = deployment;
//	}

}
