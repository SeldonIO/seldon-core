package io.seldon.apife.predictors;

import com.fasterxml.jackson.annotation.JsonProperty;

public enum InternalEndpointType {
	@JsonProperty("REST")
	REST, 
	@JsonProperty("RPC")
	RPC
}
