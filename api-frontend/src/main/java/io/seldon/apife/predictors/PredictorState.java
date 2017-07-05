package io.seldon.apife.predictors;

import java.util.Map;

public class PredictorState {
	
	public PredictiveUnitState rootState;
	public String rootId;
	public Map<String,PredictiveUnitState> predictiveUnitStatesMap;
	public Boolean enabled;
	
	public PredictorState(String rootId, Map<String,PredictiveUnitState> predictiveUnitStatesMap, Boolean enabled){
		this.rootId = rootId;
		this.predictiveUnitStatesMap = predictiveUnitStatesMap;
		this.enabled = enabled;
		this.rootState = predictiveUnitStatesMap.get(rootId);
	}
	
	public PredictiveUnitState getPredictiveUnitStatesMap(String nodeId){
		return predictiveUnitStatesMap.get(nodeId);
	}
}
