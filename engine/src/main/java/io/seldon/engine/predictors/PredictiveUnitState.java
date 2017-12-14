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
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitMethod;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.Parameter;

@JsonIgnoreProperties({"children","cluster_resources","id","subtype","type"})
public class PredictiveUnitState {
	public String name;
	public Endpoint endpoint;
	public List<PredictiveUnitState> children = new ArrayList<>();
	public Map<String,PredictiveUnitParameterInterface>  parameters;
	public String imageName;
	public String imageVersion;
	public PredictiveUnitType type;
	public PredictiveUnitImplementation implementation;
	public List<PredictiveUnitMethod> methods;
	
	@Autowired
	public PredictorBean predictorBean;
	
	public PredictiveUnitState(){}
	
	public PredictiveUnitState(
			String name,
			Endpoint endpoint,
			List<PredictiveUnitState> children,
			Map<String,PredictiveUnitParameterInterface> parameters,
			String imageName,
			String imageVersion,
			PredictiveUnitType type,
			PredictiveUnitImplementation implementation
			){
		this.name = name;
		this.endpoint = endpoint;
		this.children = children;
		this.parameters = parameters;
		this.imageName = imageName;
		this.imageVersion = imageVersion;
		this.type = type;
		this.implementation = implementation;
		
	}
	
	public PredictiveUnitState(
			PredictiveUnit predictiveUnit, 
			Map<String,Container> containersMap){
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
			this.children.add(new PredictiveUnitState(childUnit,containersMap));
		}
		
		this.type = predictiveUnit.getType();
		this.implementation = predictiveUnit.getImplementation();
		this.methods = predictiveUnit.getMethodsList();
	}
	
	public static Map<String,PredictiveUnitParameterInterface> deserializeParameters(List<Parameter> Parameters){
		Map<String,PredictiveUnitParameterInterface> paramsMap = new HashMap<>();
		for (Parameter Parameter : Parameters){
			paramsMap.put(Parameter.getName(), PredictiveUnitParameter.fromParameter(Parameter));
		}
		return paramsMap;
	}
	
	public void addChild(PredictiveUnitState predictiveUnitState){
		this.children.add(predictiveUnitState);
	}
	
}
 