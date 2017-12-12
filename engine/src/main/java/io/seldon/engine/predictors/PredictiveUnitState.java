package io.seldon.engine.predictors;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.springframework.beans.factory.annotation.Autowired;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import io.kubernetes.client.proto.V1.Container;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.Parameter;

@JsonIgnoreProperties({"children","cluster_resources","id","subtype","type"})
public class PredictiveUnitState {
	public String name;
	public PredictiveUnitBean predictiveUnitBean;
	public Endpoint endpoint;
	public List<PredictiveUnitState> children = new ArrayList<>();
	public Map<String,PredictiveUnitParameterInterface>  parameters;
	public String imageName;
	public String imageVersion;
	public PredictiveUnitType type;
	
	@Autowired
	public PredictorBean predictorBean;
	
	public PredictiveUnitState(){}
	
	public PredictiveUnitState(
			String name,
			PredictiveUnitBean predictiveUnitBean,
			Endpoint endpoint,
			List<PredictiveUnitState> children,
			Map<String,PredictiveUnitParameterInterface> parameters,
			String imageName,
			String imageVersion,
			PredictiveUnitType type
			){
		this.name = name;
		this.predictiveUnitBean = predictiveUnitBean;
		this.endpoint = endpoint;
		this.children = children;
		this.parameters = parameters;
		this.imageName = imageName;
		this.imageVersion = imageVersion;
		this.type = type;
		
	}
	
	public PredictiveUnitState(
			PredictiveUnit predictiveUnit, 
			Map<String,Container> containersMap, 
			Map<PredictiveUnitType,PredictiveUnitBean> beanTypeMap,
			Map<PredictiveUnitImplementation,PredictiveUnitBean> beanImplementationMap){
		this.name = predictiveUnit.getName();
		this.endpoint = predictiveUnit.getEndpoint();
		this.parameters = deserializeParameters(predictiveUnit.getParametersList());
		
		if (containersMap.containsKey(name)){
			String image = containersMap.get(name).getImage();
			String[] parts = image.split(":");
			this.imageName = parts[0];
			this.imageVersion = parts[1];
		}
		
		this.children = new ArrayList<PredictiveUnitState>();
		
		for (PredictiveUnit childUnit : predictiveUnit.getChildrenList()){
			this.children.add(new PredictiveUnitState(childUnit,containersMap,beanTypeMap,beanImplementationMap));
		}
		
		if ( predictiveUnit.hasImplementation()) {
			this.predictiveUnitBean = beanImplementationMap.get(predictiveUnit.getImplementation());
		}
		else {
			this.predictiveUnitBean = beanTypeMap.get(predictiveUnit.getType());
		}
		this.type = predictiveUnit.getType();
	}
	
	public static Map<String,PredictiveUnitParameterInterface> deserializeParameters(List<Parameter> Parameters){
		Map<String,PredictiveUnitParameterInterface> paramsMap = new HashMap<>();
		for (Parameter Parameter : Parameters){
			paramsMap.put(Parameter.getName(), PredictiveUnitParameter.fromParameter(Parameter));
		}
		return paramsMap;
	}
	
	public void setPredictiveUnitBean(PredictiveUnitBean predictiveUnitBean){
		this.predictiveUnitBean = predictiveUnitBean;
	}
	
	public void addChild(PredictiveUnitState predictiveUnitState){
		this.children.add(predictiveUnitState);
	}
	
}
 