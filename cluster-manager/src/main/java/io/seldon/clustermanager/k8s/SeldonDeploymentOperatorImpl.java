/*******************************************************************************
 * Copyright 2017 Seldon Technologies Ltd (http://www.seldon.io/)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *******************************************************************************/
package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.Base64;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.StringJoiner;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;

import io.kubernetes.client.models.V1OwnerReference;
import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.kubernetes.client.proto.Meta.OwnerReference;
import io.kubernetes.client.proto.V1;
import io.kubernetes.client.proto.V1.ContainerPort;
import io.kubernetes.client.proto.V1.EnvVar;
import io.kubernetes.client.proto.V1.ExecAction;
import io.kubernetes.client.proto.V1.HTTPGetAction;
import io.kubernetes.client.proto.V1.Handler;
import io.kubernetes.client.proto.V1.Lifecycle;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.kubernetes.client.proto.V1.Probe;
import io.kubernetes.client.proto.V1.Service;
import io.kubernetes.client.proto.V1.ServicePort;
import io.kubernetes.client.proto.V1.ServiceSpec;
import io.kubernetes.client.proto.V1.TCPSocketAction;
import io.kubernetes.client.proto.V1beta1Extensions;
import io.kubernetes.client.proto.V1beta1Extensions.Deployment;
import io.kubernetes.client.proto.V1beta1Extensions.DeploymentSpec;
import io.kubernetes.client.proto.V1beta1Extensions.DeploymentStrategy;
import io.kubernetes.client.proto.V1beta1Extensions.RollingUpdateDeployment;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.Endpoint;
import io.seldon.protos.DeploymentProtos.Parameter;
import io.seldon.protos.DeploymentProtos.PredictiveUnit;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitSubtype;
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitType;
import io.seldon.protos.DeploymentProtos.PredictorSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

@Component
public class SeldonDeploymentOperatorImpl implements SeldonDeploymentOperator {

	private final static Logger logger = LoggerFactory.getLogger(SeldonDeploymentOperatorImpl.class);
	private final ClusterManagerProperites clusterManagerProperites;
	public static final String LABEL_SELDON_APP = "seldon-app";
    public static final String LABEL_SELDON_TYPE_KEY = "seldon-type";
    public static final String LABEL_SELDON_TYPE_VAL = "deployment";
   
	@Autowired
	public SeldonDeploymentOperatorImpl(ClusterManagerProperites clusterManagerProperites) {
		super();
		this.clusterManagerProperites = clusterManagerProperites;
	}

	 
	private static String getEngineEnvVarJson(Message protoMessage) throws SeldonDeploymentException {
		String retVal;
		try {
            retVal = ProtoBufUtils.toJson(protoMessage, true,false);
            retVal = new String(Base64.getEncoder().encode(retVal.getBytes()));
            return retVal;
		} catch (InvalidProtocolBufferException e) {
           throw new SeldonDeploymentException("Failed to parse protobuf",e);
        }
	}
	
	private V1.Container createEngineContainer(SeldonDeployment dep,PredictorSpec predictorDef) throws SeldonDeploymentException
	{
		V1.Container.Builder cBuilder = V1.Container.newBuilder();
		cBuilder
			.setName("seldon-container-engine")
			.setImage(clusterManagerProperites.getEngineContainerImageAndVersion())
			.addEnv(EnvVar.newBuilder().setName("ENGINE_PREDICTOR").setValue(getEngineEnvVarJson(predictorDef)))
			.addEnv(EnvVar.newBuilder().setName("ENGINE_SELDON_DEPLOYMENT").setValue(getEngineEnvVarJson(dep)))
			.addEnv(EnvVar.newBuilder().setName("ENGINE_SERVER_PORT").setValue(""+clusterManagerProperites.getEngineContainerPort()))
			.addEnv(EnvVar.newBuilder().setName("ENGINE_SERVER_GRPC_PORT").setValue(""+clusterManagerProperites.getEngineGrpcContainerPort()))			
			.addPorts(V1.ContainerPort.newBuilder().setContainerPort(clusterManagerProperites.getEngineContainerPort()))
			.addPorts(V1.ContainerPort.newBuilder().setContainerPort(8082).setName("admin"))
			.setReadinessProbe(Probe.newBuilder().setHandler(Handler.newBuilder()
					.setHttpGet(HTTPGetAction.newBuilder().setPort(IntOrString.newBuilder().setType(1).setStrVal("admin")).setPath("/ready")))
					.setInitialDelaySeconds(10)
					.setPeriodSeconds(10)
					.setFailureThreshold(3)
					.setSuccessThreshold(1)
					.setTimeoutSeconds(2)
					)
			.setLivenessProbe(Probe.newBuilder().setHandler(Handler.newBuilder()
					.setHttpGet(HTTPGetAction.newBuilder().setPort(IntOrString.newBuilder().setType(1).setStrVal("admin")).setPath("/ready")))
					.setInitialDelaySeconds(10)
					.setPeriodSeconds(10)
					.setFailureThreshold(3)
					.setSuccessThreshold(1)
					.setTimeoutSeconds(2)
					)
			.setLifecycle(Lifecycle.newBuilder().setPreStop(Handler.newBuilder().setExec(
					ExecAction.newBuilder()
						.addCommand("/bin/sh")
						.addCommand("-c")
						.addCommand("curl 127.0.0.1:"+clusterManagerProperites.getEngineContainerPort()+"/pause && /bin/sleep 5"))));

			
			
		return cBuilder.build();
	}
	
	
	
	private Set<String> getEnvNamesProto(List<EnvVar> envs)
	{
		Set<String> s = new HashSet<>();
		for(EnvVar e : envs)
			s.add(e.getName());
		return s;
	}
	
	private Integer getPort(List<ContainerPort> ports)
	{
	    if (ports != null)
	        for(ContainerPort p : ports)
	            if ("http".equals(p.getName()) || "grpc".equals(p.getName()))
	                return p.getContainerPort();
	    return null;
	}
	
	private String extractPredictiveUnitParametersAsJson(PredictiveUnit predictiveUnit) {
	    if (predictiveUnit == null)
	        return "";
        StringJoiner sj = new StringJoiner(",", "[", "]");
        List<Parameter> parameters = predictiveUnit.getParametersList();
        for (Parameter parameter : parameters) {
            try {
                String j = ProtoBufUtils.toJson(parameter, true,false);
                sj.add(j);
            } catch (InvalidProtocolBufferException e) {
                throw new RuntimeException(e);
            }
        }
        return sj.toString();
    }
	
	private PredictiveUnit findPredictiveUnitForContainer(PredictiveUnit unit,String name)
	{
	    if (unit.getName().equals(name))
	        return unit;
	    else {
	        for(PredictiveUnit child : unit.getChildrenList())
	        {
	            PredictiveUnit found = findPredictiveUnitForContainer(child,name);
	            if (found != null)
	                return found;
	        }
	        return null;
	    }
	}
	
	private V1.Container updateContainer(V1.Container c,PredictiveUnit pu,int idx)
	{
		V1.Container.Builder c2Builder = V1.Container.newBuilder(c);
        
		Integer containerPort = getPort(c.getPortsList());
		if (containerPort == null)
		{
		    if (pu != null)
		    {
		        if (pu.getEndpoint().getType() == Endpoint.EndpointType.REST)
		        {
		            c2Builder.addPorts(ContainerPort.newBuilder().setName("http").setContainerPort(clusterManagerProperites.getPuContainerPortBase() + idx));
		            containerPort = clusterManagerProperites.getPuContainerPortBase() + idx;
		        }
		        else
		        {
		            c2Builder.addPorts(ContainerPort.newBuilder().setName("grpc").setContainerPort(clusterManagerProperites.getPuContainerPortBase() + idx));
		            containerPort = clusterManagerProperites.getPuContainerPortBase() + idx;		        
		        }
		    }
		}
		else
			containerPort = c.getPorts(0).getContainerPort();
		
		final String ENV_PREDICTIVE_UNIT_SERVICE_PORT ="PREDICTIVE_UNIT_SERVICE_PORT";
		Set<String> envNames = this.getEnvNamesProto(c.getEnvList());
		if (!envNames.contains(ENV_PREDICTIVE_UNIT_SERVICE_PORT))
			c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_PREDICTIVE_UNIT_SERVICE_PORT).setValue(""+containerPort));
				
		final String ENV_PREDICTIVE_UNIT_PARAMETERS = "PREDICTIVE_UNIT_PARAMETERS";
		if (!envNames.contains(ENV_PREDICTIVE_UNIT_PARAMETERS))
		    c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_PREDICTIVE_UNIT_PARAMETERS).setValue(extractPredictiveUnitParametersAsJson(pu)));
		
		if (!c.hasLivenessProbe())
		{
			c2Builder.setLivenessProbe(Probe.newBuilder()
					.setHandler(Handler.newBuilder().setTcpSocket(TCPSocketAction.newBuilder().setPort(io.kubernetes.client.proto.IntStr.IntOrString.newBuilder().setType(1).setStrVal("http"))))
					.setInitialDelaySeconds(10)
					.setPeriodSeconds(5)
					);
		}
		
		if (!c.hasReadinessProbe())
		{
			c2Builder.setReadinessProbe(Probe.newBuilder()
					.setHandler(Handler.newBuilder().setTcpSocket(TCPSocketAction.newBuilder().setPort(io.kubernetes.client.proto.IntStr.IntOrString.newBuilder().setType(1).setStrVal("http"))))
					.setInitialDelaySeconds(10)
					.setPeriodSeconds(5)
					);
			
		}
		
		
		if (!c.hasLifecycle())
		{
			if (!c.getLifecycle().hasPreStop())
			{
				c2Builder.setLifecycle(Lifecycle.newBuilder(c.getLifecycle()).setPreStop(Handler.newBuilder().setExec(
						ExecAction.newBuilder().addCommand("/bin/sh").addCommand("-c").addCommand("/bin/sleep 5"))));
			}
		}

		
		return c2Builder.build();
	}
	
	private void updatePredictiveUnitBuilderByName(PredictiveUnit.Builder puBuilder,V1.Container container)
	{
	    if (puBuilder.getName().equals(container.getName()))
        {
            Endpoint.Builder b = puBuilder.getEndpointBuilder();
            for(ContainerPort p : container.getPortsList())
            {
                if ("http".equals(p.getName())) //first found will be used
                {
                    b.setServicePort(p.getContainerPort());
                    b.setType(Endpoint.EndpointType.REST);
                    b.setServiceHost("0.0.0.0"); //assumes localhost at present
                    return;
                } else if ("grpc".equals(p.getName())) {
                    b.setServicePort(p.getContainerPort());
                    b.setType(Endpoint.EndpointType.GRPC);
                    b.setServiceHost("0.0.0.0"); //assumes localhost at present                    
                    return;
                }
            }
        } else {
            for(int i=0;i<puBuilder.getChildrenCount();i++)
	            updatePredictiveUnitBuilderByName(puBuilder.getChildrenBuilder(i),container);
        }
	}
	
	@Override
	public SeldonDeployment defaulting(SeldonDeployment mlDep) {
		SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder(mlDep);
		int idx = 0;
		String serviceName = mlDep.getSpec().getName();
		for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
		{
			ObjectMeta.Builder metaBuilder = ObjectMeta.newBuilder(p.getComponentSpec().getMetadata())
				.putLabels(LABEL_SELDON_APP, serviceName);
			mlBuilder.getSpecBuilder().getPredictorsBuilder(idx).getComponentSpecBuilder().setMetadata(metaBuilder);
			int cIdx = 0;
			mlBuilder.getSpecBuilder().getPredictorsBuilder(idx).getComponentSpecBuilder().getSpecBuilder().clearContainers();
			for(V1.Container c : p.getComponentSpec().getSpec().getContainersList())
			{
				V1.Container c2 = this.updateContainer(c, findPredictiveUnitForContainer(mlDep.getSpec().getPredictors(idx).getGraph(),c.getName()),cIdx);
				mlBuilder.getSpecBuilder().getPredictorsBuilder(idx).getComponentSpecBuilder().getSpecBuilder().addContainers(cIdx, c2);	
				updatePredictiveUnitBuilderByName(mlBuilder.getSpecBuilder().getPredictorsBuilder(idx).getGraphBuilder(),c2);
				cIdx++;
			}
			idx++;
		}
		
		return mlBuilder.build();
	}
	
	
	private void checkPredictiveUnitsMicroservices(PredictiveUnit pu,PredictorSpec p) throws SeldonDeploymentException
	{
        if (pu.hasType() &&
                pu.hasSubtype() && 
                pu.getType() == PredictiveUnitType.MODEL && 
                pu.getSubtype() == PredictiveUnitSubtype.MICROSERVICE)
        {
            boolean found = false;
            for(V1.Container c : p.getComponentSpec().getSpec().getContainersList())
            {
                if (c.getName().equals(pu.getName()))
                {
                    found = true;
                    break;
                }
            }
            if (!found)
            {
                throw new SeldonDeploymentException("Can't find container for predictive unit with name "+pu.getName());    
            }
        }
        for(PredictiveUnit child :  pu.getChildrenList())
            checkPredictiveUnitsMicroservices(child,p);
	}
	
	private void checkTypeAndSubType(PredictiveUnit pu) throws SeldonDeploymentException
	{
        if (!pu.hasType())
            throw new SeldonDeploymentException(String.format("Predictive unit %s has no type",pu.getName()));    
        if (!pu.hasSubtype())
            throw new SeldonDeploymentException(String.format("Predictive unit %s has no subtype",pu.getName()));  
        for(PredictiveUnit child :  pu.getChildrenList())
            checkTypeAndSubType(child); 
	}

	@Override
	public void validate(SeldonDeployment mlDep) throws SeldonDeploymentException {

	    for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
        {
	        checkPredictiveUnitsMicroservices(p.getGraph(),p);
	        checkTypeAndSubType(p.getGraph());
        }
        
	}
	
	@Override
	public String getKubernetesDeploymentName(String deploymentName,String predictorName) {
		return deploymentName + "-" + predictorName;
	}
	
	private V1OwnerReference getOwnerReferenceOld(SeldonDeployment mlDep)
	{
		return new V1OwnerReference()
				.apiVersion(mlDep.getApiVersion())
				.kind(mlDep.getKind())
				.controller(true)
				.name(mlDep.getMetadata().getName())
				.uid(mlDep.getMetadata().getUid());
	}
	private OwnerReference getOwnerReference(SeldonDeployment mlDep)
	{
		return OwnerReference.newBuilder()
			.setApiVersion(mlDep.getApiVersion())
			.setKind(mlDep.getKind())
			.setController(true)
			.setName(mlDep.getMetadata().getName())
			.setUid(mlDep.getMetadata().getUid()).build();
	}
	 
	@Override
	public DeploymentResources createResources(SeldonDeployment mlDep) throws SeldonDeploymentException {
		
		OwnerReference ownerRef = getOwnerReference(mlDep);
		List<Deployment> deployments = new ArrayList<>();
		// for each predictor Create/replace deployment
		String serviceLabel = mlDep.getSpec().getName();
		for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
		{
			String depName = getKubernetesDeploymentName(mlDep.getSpec().getName(),p.getName());
			PodTemplateSpec.Builder podSpecBuilder = PodTemplateSpec.newBuilder(p.getComponentSpec());
			podSpecBuilder.getSpecBuilder()
			    .addContainers(createEngineContainer(mlDep,p))
			    .setTerminationGracePeriodSeconds(20);
			podSpecBuilder.getMetadataBuilder()
			    .putAnnotations("prometheus.io/path", "/prometheus")
			    .putAnnotations("prometheus.io/port",""+clusterManagerProperites.getEngineContainerPort())
			    .putAnnotations("prometheus.io/scrape", "true");

			Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
					.setMetadata(ObjectMeta.newBuilder()
							.setName(depName)
							.putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel)
							.putLabels(Constants.LABEL_SELDON_ID, mlDep.getSpec().getName())
							.putLabels("app", depName)
							.putLabels("version", "v1") //FIXME
							.putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_KEY, SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_VAL)
							.addOwnerReferences(ownerRef)
							)
					.setSpec(DeploymentSpec.newBuilder()
					        .setTemplate(podSpecBuilder.build())
					        .setStrategy(DeploymentStrategy.newBuilder().setRollingUpdate(RollingUpdateDeployment.newBuilder().setMaxUnavailable(IntOrString.newBuilder().setType(1).setStrVal("10%"))))
							.setReplicas(p.getReplicas()))
					.build();
			
			deployments.add(deployment);
		}
		
		Service s = Service.newBuilder()
					.setMetadata(ObjectMeta.newBuilder()
							.setName(mlDep.getSpec().getName())
							.putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel)
							.putLabels("seldon-deployment-id", mlDep.getSpec().getName())
							.addOwnerReferences(ownerRef)
							)
					.setSpec(ServiceSpec.newBuilder()
                            .addPorts(ServicePort.newBuilder()
                                    .setProtocol("TCP")
                                    .setPort(clusterManagerProperites.getEngineContainerPort())
                                    .setTargetPort(IntOrString.newBuilder().setIntVal(clusterManagerProperites.getEngineContainerPort()))
                                    .setName("http")
                                    )
                            .addPorts(ServicePort.newBuilder()
                                    .setProtocol("TCP")
                                    .setPort(clusterManagerProperites.getEngineGrpcContainerPort())
                                    .setTargetPort(IntOrString.newBuilder().setIntVal(clusterManagerProperites.getEngineGrpcContainerPort()))
                                    .setName("grpc")
                                    )
							.setType("ClusterIP")
							.putSelector(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP,serviceLabel)
							)
				.build();
		
		// Create service for deployment
		return new DeploymentResources(deployments, s);
	}
	

	public static class DeploymentResources {
		
		List<Deployment> deployments;
		Service service;
		
		public DeploymentResources(List<Deployment> deployments, Service service) {
			super();
			this.deployments = deployments;
			this.service = service;
		}
		

	}


	
}
