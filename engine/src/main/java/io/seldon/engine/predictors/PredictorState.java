package io.seldon.engine.predictors;

import java.util.Map;

public class PredictorState {
	
	public PredictiveUnitState rootState;
	public Long rootId;
	public Map<Long,PredictiveUnitState> predictiveUnitStatesMap;
	public Boolean enabled;
	
	public PredictorState(Long rootId, Map<Long,PredictiveUnitState> predictiveUnitStatesMap, Boolean enabled){
		this.rootId = rootId;
		this.predictiveUnitStatesMap = predictiveUnitStatesMap;
		this.enabled = enabled;
		this.rootState = predictiveUnitStatesMap.get(rootId);
	}
	
	public PredictiveUnitState getPredictiveUnitStatesMap(String nodeId){
		return predictiveUnitStatesMap.get(nodeId);
	}
}
