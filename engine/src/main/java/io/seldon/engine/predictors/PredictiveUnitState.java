package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;

@JsonIgnoreProperties({"children","cluster_resources","id","subtype","type"})
public class PredictiveUnitState {
	public String id;
	public String name;
	public PredictiveUnitBean predictiveUnitBean;
	public EndpointDef endpoint;
	public List<PredictiveUnitState> children = new ArrayList<>();
	public Map<String,PredictiveUnitParameterInterface>  parameters;
	
	public PredictiveUnitState(){}
	
	public PredictiveUnitState(String id,String name, EndpointDef endpoint, Map<String,PredictiveUnitParameterInterface> parameters){
		this.id = id;
		this.name = name;
		this.endpoint = endpoint;
		this.parameters = parameters;
	}
	
	public PredictiveUnitState(String id,String name, 
			PredictiveUnitBean predictiveUnitBean, 
			EndpointDef endpoint, 
			List<PredictiveUnitState> children, 
			Map<String,PredictiveUnitParameterInterface> parameters){
		this(	id,
				name,
				endpoint,
				parameters
				);
		this.predictiveUnitBean = predictiveUnitBean;
		this.children = children;
	}
	
	public PredictiveUnitState(PredictiveUnitDef predictiveUnitDef){
		this(	predictiveUnitDef.getId(),
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
		this.children.add(child);
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
 