package io.seldon.engine.predictors;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;

@JsonIgnoreProperties({"children","cluster_resources","id","subtype","type"})
public class PredictiveUnitState {
	public String name;
	public PredictiveUnitBean predictiveUnitBean;
	public EndpointDef endpoint;
	public Map<String,PredictiveUnitState> children = new HashMap<>();
	public Map<String,PredictiveUnitParameterInterface>  parameters;
	
	public PredictiveUnitState(){}
	
	public PredictiveUnitState(String name, EndpointDef endpoint, Map<String,PredictiveUnitParameterInterface> parameters){
		this.name = name;
		this.endpoint = endpoint;
		this.parameters = parameters;
	}
	
	public PredictiveUnitState(String name, 
			PredictiveUnitBean predictiveUnitBean, 
			EndpointDef endpoint, 
			Map<String,PredictiveUnitState> children, 
			Map<String,PredictiveUnitParameterInterface> parameters){
		this(
				name,
				endpoint,
				parameters
				);
		this.predictiveUnitBean = predictiveUnitBean;
		this.children = children;
	}
	
	public PredictiveUnitState(PredictiveUnitDef predictiveUnitDef){
		this(
				predictiveUnitDef.getName(),
				predictiveUnitDef.getEndpoint(),
				deserializeParameters(predictiveUnitDef.getParametersList())
				);
	}
	
	public static Map<String,PredictiveUnitParameterInterface> deserializeParameters(List<ParamDef> paramDefs){
		Map<String,PredictiveUnitParameterInterface> paramsMap = new HashMap<>();
		for (ParamDef paramDef : paramDefs){
			paramsMap.put(paramDef.getName(), PredictiveUnitParameter.fromParamDef(paramDef));
		}
		return paramsMap;
	}
	
	public void addChild(String childId,PredictiveUnitState child){
		this.children.put(childId, child);
	}
	
	public void setPredictiveUnitBean(PredictiveUnitBean predictiveUnitBean){
		this.predictiveUnitBean = predictiveUnitBean;
	}
	
//	public static PredictiveUnitState fromPredictiveUnitDef(PredictiveUnitDef predictiveUnitDef){
//		String name = predictiveUnitDef.getName();
//		EndpointDef endpoint = predictiveUnitDef.getEndpoint();
//		List<ParamDef> parameters = predictiveUnitDef.getParametersList();
//		
//		return new PredictiveUnitState(name,endpoint,parameters);
//	}
}
 