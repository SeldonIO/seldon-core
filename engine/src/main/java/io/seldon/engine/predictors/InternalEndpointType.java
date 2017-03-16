package io.seldon.engine.predictors;

import com.fasterxml.jackson.annotation.JsonProperty;

public enum InternalEndpointType {
	@JsonProperty("REST")
	REST, 
	@JsonProperty("RPC")
	RPC
}
