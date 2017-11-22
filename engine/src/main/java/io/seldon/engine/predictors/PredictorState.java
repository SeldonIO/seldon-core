package io.seldon.engine.predictors;

import java.util.Map;
import io.kubernetes.client.proto.V1.Container;

public class PredictorState {
	
	public PredictiveUnitState rootState;
	public String rootName;
	public Boolean enabled;
	
	public PredictorState(String rootName, PredictiveUnitState rootState, Boolean enabled){
		this.rootName = rootName;
		this.enabled = enabled;
		this.rootState = rootState;
	}
	
	public PredictiveUnitState getRootState(){
		return this.rootState;
	}
}
