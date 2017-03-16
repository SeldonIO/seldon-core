package io.seldon.engine.predictors;

import com.fasterxml.jackson.annotation.JsonProperty;

public enum PredictiveUnitParameterType {
	@JsonProperty("STR")
	STR, 
	@JsonProperty("FLOAT")
	FLOAT, 
	@JsonProperty("INT")
	INT
}
