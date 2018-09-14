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
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.StringJoiner;

import org.apache.commons.codec.digest.DigestUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;

import io.kubernetes.client.models.V1OwnerReference;
import io.kubernetes.client.proto.IntStr.IntOrString;
import io.kubernetes.client.proto.Meta.LabelSelector;
import io.kubernetes.client.proto.Meta.LabelSelectorRequirement;
import io.kubernetes.client.proto.Meta.ObjectMeta;
import io.kubernetes.client.proto.Meta.OwnerReference;
import io.kubernetes.client.proto.Resource.Quantity;
import io.kubernetes.client.proto.V1;
import io.kubernetes.client.proto.V1.ContainerPort;
import io.kubernetes.client.proto.V1.DownwardAPIVolumeFile;
import io.kubernetes.client.proto.V1.DownwardAPIVolumeSource;
import io.kubernetes.client.proto.V1.EnvVar;
import io.kubernetes.client.proto.V1.ExecAction;
import io.kubernetes.client.proto.V1.HTTPGetAction;
import io.kubernetes.client.proto.V1.Handler;
import io.kubernetes.client.proto.V1.Lifecycle;
import io.kubernetes.client.proto.V1.ObjectFieldSelector;
import io.kubernetes.client.proto.V1.PodSecurityContext;
import io.kubernetes.client.proto.V1.PodTemplateSpec;
import io.kubernetes.client.proto.V1.Probe;
import io.kubernetes.client.proto.V1.SecurityContext;
import io.kubernetes.client.proto.V1.Service;
import io.kubernetes.client.proto.V1.ServicePort;
import io.kubernetes.client.proto.V1.ServiceSpec;
import io.kubernetes.client.proto.V1.TCPSocketAction;
import io.kubernetes.client.proto.V1.Volume;
import io.kubernetes.client.proto.V1.VolumeMount;
import io.kubernetes.client.proto.V1.VolumeSource;
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
import io.seldon.protos.DeploymentProtos.PredictiveUnit.PredictiveUnitImplementation;
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
    public static final String PODINFO_VOLUME_NAME = "podinfo";
    public static final String PODINFO_VOLUME_PATH = "/etc/podinfo";
   
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
	
	static final String ENGINE_JAVA_OPTS="-Dcom.sun.management.jmxremote.rmi.port=9090 -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=9090 -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.local.only=false -Djava.rmi.server.hostname=127.0.0.1";
	private V1.Container createEngineContainer(SeldonDeployment mlDep,PredictorSpec predictorDef) throws SeldonDeploymentException
	{
		V1.Container.Builder cBuilder = V1.Container.newBuilder();
		
		cBuilder
			.setName("seldon-container-engine")
			.setImage(clusterManagerProperites.getEngineContainerImageAndVersion())
			.setImagePullPolicy(clusterManagerProperites.getEngineContainerImagePullPolicy())
			.addVolumeMounts(VolumeMount.newBuilder().setName(PODINFO_VOLUME_NAME).setMountPath(PODINFO_VOLUME_PATH).setReadOnly(true))
			.addEnv(EnvVar.newBuilder().setName("ENGINE_PREDICTOR").setValue(getEngineEnvVarJson(predictorDef)))
			.addEnv(EnvVar.newBuilder().setName("DEPLOYMENT_NAME").setValue(mlDep.getSpec().getName()))		
			.addEnv(EnvVar.newBuilder().setName("ENGINE_SERVER_PORT").setValue(""+clusterManagerProperites.getEngineContainerPort()))
			.addEnv(EnvVar.newBuilder().setName("ENGINE_SERVER_GRPC_PORT").setValue(""+clusterManagerProperites.getEngineGrpcContainerPort()))			
			.addEnv(EnvVar.newBuilder().setName("JAVA_OPTS").setValue(ENGINE_JAVA_OPTS))						
			.addPorts(V1.ContainerPort.newBuilder().setContainerPort(clusterManagerProperites.getEngineContainerPort()))
			.addPorts(V1.ContainerPort.newBuilder().setContainerPort(8082).setName("admin"))
			.addPorts(V1.ContainerPort.newBuilder().setContainerPort(9090).setName("jmx"))
			.setSecurityContext(SecurityContext.newBuilder().setRunAsUser(8888).build())
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

		// Add engine resources if specified
		if (predictorDef.hasEngineResources())
		    cBuilder.setResources(predictorDef.getEngineResources());
		else {// set default resource requests for cpu
			final String DEFAULT_ENGINE_CPU_REQUEST = "0.1";
		    cBuilder.setResources(V1.ResourceRequirements.newBuilder().putRequests("cpu", Quantity.newBuilder().setString(DEFAULT_ENGINE_CPU_REQUEST).build()));
		}
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
	
	private V1.Container updateContainer(V1.Container c,PredictiveUnit pu,int portNum,String deploymentName,String predictorName)
	{
		V1.Container.Builder c2Builder = V1.Container.newBuilder(c);
        
		//Add volume to get at pod annotations
		c2Builder.addVolumeMounts(VolumeMount.newBuilder().setName(PODINFO_VOLUME_NAME).setMountPath(PODINFO_VOLUME_PATH).setReadOnly(true));
		
		Integer containerPort = getPort(c.getPortsList());
		// Add container port and liveness and readiness probes if no container ports are specified
		if (containerPort == null)
		{
		    if (pu != null)
		    {
		        if (pu.getEndpoint().getType() == Endpoint.EndpointType.REST)
		        {
		            c2Builder.addPorts(ContainerPort.newBuilder().setName("http").setContainerPort(portNum));
		            containerPort = portNum;
		            
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
		        }
		        else
		        {
		            c2Builder.addPorts(ContainerPort.newBuilder().setName("grpc").setContainerPort(portNum));
		            containerPort = portNum;		
		            
		            if (!c.hasLivenessProbe())
                    {
                        c2Builder.setLivenessProbe(Probe.newBuilder()
                                .setHandler(Handler.newBuilder().setTcpSocket(TCPSocketAction.newBuilder().setPort(io.kubernetes.client.proto.IntStr.IntOrString.newBuilder().setType(1).setStrVal("grpc"))))
                                .setInitialDelaySeconds(10)
                                .setPeriodSeconds(5)
                                );
                    }
                    if (!c.hasReadinessProbe())
                    {
                        
                        c2Builder.setReadinessProbe(Probe.newBuilder()
                                .setHandler(Handler.newBuilder().setTcpSocket(TCPSocketAction.newBuilder().setPort(io.kubernetes.client.proto.IntStr.IntOrString.newBuilder().setType(1).setStrVal("grpc"))))
                                .setInitialDelaySeconds(10)
                                .setPeriodSeconds(5)
                                );
                        
                    }
		        }
		    }
		}
		else
		{
			//throw new UnsupportedOperationException(String.format("Found container port already set with http or grpc label. This is not presently allowed. Found port {}",containerPort));
		}
		
		// Add environment variable for the port used in case the model needs to access it
		final String ENV_PREDICTIVE_UNIT_SERVICE_PORT ="PREDICTIVE_UNIT_SERVICE_PORT";
		Set<String> envNames = this.getEnvNamesProto(c.getEnvList());
		if (!envNames.contains(ENV_PREDICTIVE_UNIT_SERVICE_PORT))
			c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_PREDICTIVE_UNIT_SERVICE_PORT).setValue(""+containerPort));
				
		//Add environment variable for the parameters passed in case the model needs to access it
		final String ENV_PREDICTIVE_UNIT_PARAMETERS = "PREDICTIVE_UNIT_PARAMETERS";
		if (!envNames.contains(ENV_PREDICTIVE_UNIT_PARAMETERS))
		    c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_PREDICTIVE_UNIT_PARAMETERS).setValue(extractPredictiveUnitParametersAsJson(pu)));

		//Add environment variable for the predictive unit ID, the predictor ID and the Deployment ID
		final String ENV_PREDICTIVE_UNIT_ID = "PREDICTIVE_UNIT_ID";
		final String ENV_PREDICTOR_ID = "PREDICTOR_ID";
		final String ENV_SELDON_DEPLOYMENT_ID = "SELDON_DEPLOYMENT_ID";
		if (!envNames.contains(ENV_PREDICTIVE_UNIT_ID))
			c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_PREDICTIVE_UNIT_ID).setValue(c.getName()));
		if (!envNames.contains(ENV_PREDICTOR_ID))
			c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_PREDICTOR_ID).setValue(predictorName));
		if (!envNames.contains(ENV_SELDON_DEPLOYMENT_ID))
			c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_SELDON_DEPLOYMENT_ID).setValue(deploymentName));
		
		// Add a default lifecycle pre-stop if non exists
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
	
	private void updatePredictiveUnitBuilderByName(PredictiveUnit.Builder puBuilder,V1.Container container,String containerHostName)
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
                    b.setServiceHost(containerHostName);
                    return;
                } else if ("grpc".equals(p.getName())) {
                    b.setServicePort(p.getContainerPort());
                    b.setType(Endpoint.EndpointType.GRPC);
                    b.setServiceHost(containerHostName);
                    return;
                }
            }
        } else {
            for(int i=0;i<puBuilder.getChildrenCount();i++)
	            updatePredictiveUnitBuilderByName(puBuilder.getChildrenBuilder(i),container,containerHostName);
        }
	}
	
	/*
	private String getPredictorServiceNameValue(SeldonDeployment mlDep,String predictorName,String containerName)
	{
		return  mlDep.getSpec().getName()+"-"+predictorName+"-"+containerName;
	}
	*/
	private String getPredictorServiceNameKey(String containerName)
	{
		return LABEL_SELDON_APP+"-"+containerName;
	}
	
	private String hash(String key)  {
	    return DigestUtils.md5Hex(key).toLowerCase();
	}
	
	@Override
	public String getSeldonServiceName(SeldonDeployment dep,PredictorSpec pred,String key) {
		String svcName =  dep.getSpec().getName() + "-" + pred.getName()+"-"+key;
		if (svcName.length() > 63)
			return "seldon-"+hash(svcName);
		else
			return svcName;
	}
	
	
	@Override
	public SeldonDeployment defaulting(SeldonDeployment mlDep) {
		SeldonDeployment.Builder mlBuilder = SeldonDeployment.newBuilder(mlDep);
		String deploymentName = mlDep.getMetadata().getName();
		
		for(int pbIdx=0;pbIdx<mlDep.getSpec().getPredictorsCount();pbIdx++)
		{
			PredictorSpec p = mlDep.getSpec().getPredictors(pbIdx);
			Map<String,Integer> servicePortMap = new HashMap<>();
			int currentServicePortNum = clusterManagerProperites.getPuContainerPortBase();
			for(int ptsIdx=0;ptsIdx<p.getComponentSpecsCount();ptsIdx++)
			{
				V1.PodTemplateSpec spec = p.getComponentSpecs(ptsIdx);
				ObjectMeta.Builder metaBuilder = ObjectMeta.newBuilder(spec.getMetadata());

				mlBuilder.getSpecBuilder().getPredictorsBuilder(pbIdx).getComponentSpecsBuilder(ptsIdx).getSpecBuilder().clearContainers();
				String predictorName = p.getName();
				for(int cIdx = 0;cIdx < spec.getSpec().getContainersCount();cIdx++)
				{
					V1.Container c = spec.getSpec().getContainers(cIdx);
					String containerServiceKey = getPredictorServiceNameKey(c.getName());
					String containerServiceValue = getSeldonServiceName(mlDep, p, c.getName());
					metaBuilder.putLabels(containerServiceKey, containerServiceValue); 
					
					int portNum;
					if (servicePortMap.containsKey(c.getName()))
						portNum = servicePortMap.get(c.getName());
					else
					{
						portNum = currentServicePortNum;
						servicePortMap.put(c.getName(), portNum);
						currentServicePortNum++;
					}
					V1.Container c2 = this.updateContainer(c, findPredictiveUnitForContainer(mlDep.getSpec().getPredictors(pbIdx).getGraph(),c.getName()),portNum,deploymentName,predictorName);
					mlBuilder.getSpecBuilder().getPredictorsBuilder(pbIdx).getComponentSpecsBuilder(ptsIdx).getSpecBuilder().addContainers(cIdx, c2);	
					updatePredictiveUnitBuilderByName(mlBuilder.getSpecBuilder().getPredictorsBuilder(pbIdx).getGraphBuilder(),c2,containerServiceValue); 
				}
				mlBuilder.getSpecBuilder().getPredictorsBuilder(pbIdx).getComponentSpecsBuilder(ptsIdx).setMetadata(metaBuilder);
			}
		}	
		
		if (!mlBuilder.hasStatus())
		{
			mlBuilder.getStatusBuilder().setState(Constants.STATE_CREATING);
		}
		
		return mlBuilder.build();
	}
	
	
	private void checkPredictiveUnitsMicroservices(PredictiveUnit pu,PredictorSpec p) throws SeldonDeploymentException
	{
        if (pu.hasType() &&
                pu.getType() == PredictiveUnitType.MODEL &&
                pu.getImplementation() == PredictiveUnitImplementation.UNKNOWN_IMPLEMENTATION)
        {
            boolean found = false;
            for(V1.PodTemplateSpec spec : p.getComponentSpecsList())
            	for(V1.Container c : spec.getSpec().getContainersList())
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
	
	/*
	 * If implementation is specified, ignore the rest
	 * if not, implementation defaults to UNKNOWN_IMPLEMENTATION and
	 * if type is specified ignore the rest
	 * if not, type defaults to UNKNOWN_TYPE and
	 * if methods is not specified, raise an error (we are in the case when none of implementation, type, methods has been specified)
	 */
	private void checkTypeMethodAndImpl(PredictiveUnit pu) throws SeldonDeploymentException
	{
        if ((!pu.hasImplementation() || pu.getImplementation().getNumber() == PredictiveUnitImplementation.UNKNOWN_IMPLEMENTATION_VALUE) &&
                (!pu.hasType() || pu.getType().getNumber() == PredictiveUnitType.UNKNOWN_TYPE_VALUE) &&
                pu.getMethodsCount() == 0) 
            throw new SeldonDeploymentException(String.format("Predictive unit %s has no methods specified",pu.getName()));     
        for(PredictiveUnit child :  pu.getChildrenList())
            checkTypeMethodAndImpl(child); 
	}

	@Override
	public void validate(SeldonDeployment mlDep) throws SeldonDeploymentException {

	    for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
        {
	        checkPredictiveUnitsMicroservices(p.getGraph(),p);
	        checkTypeMethodAndImpl(p.getGraph());
        }
        
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
	 
	private String getAmbassadorAnnotation(SeldonDeployment mlDep,String serviceName)
	{
        final String restMapping = "---\n"+
                "apiVersion: ambassador/v0\n" +
                "kind:  Mapping\n" +
                "name:  seldon_"+mlDep.getMetadata().getName()+"_rest_mapping\n" +
                "prefix: /seldon/"+mlDep.getMetadata().getName()+"/\n" +
                "service: "+serviceName+":"+clusterManagerProperites.getEngineContainerPort()+"\n";
        final String grpcMapping = "---\n"+
                "apiVersion: ambassador/v0\n" +
                "kind:  Mapping\n" +
                "name:  "+mlDep.getMetadata().getName()+"_grpc_mapping\n" +
                "grpc: true\n" +
                "prefix: /seldon.protos.Seldon/\n" +                
                "rewrite: /seldon.protos.Seldon/\n" + 
                "headers:\n"+
                 "  seldon: "+mlDep.getMetadata().getName() + "\n" +
                "service: "+serviceName+":"+clusterManagerProperites.getEngineGrpcContainerPort()+"\n";
	    return restMapping + grpcMapping;
	}
	
	
	private void addServicePort(PredictiveUnit pu,String serviceName,ServiceSpec.Builder svcSpecBuilder)
	{
		if (pu.hasEndpoint())
		{
			Endpoint e = pu.getEndpoint();
			if (e.getServiceHost().equals(serviceName))
			{
				switch(e.getType())
				{
				case REST:
					svcSpecBuilder.addPorts(ServicePort.newBuilder()
							.setProtocol("TCP")
							.setPort(e.getServicePort())
							.setTargetPort(IntOrString.newBuilder().setIntVal(e.getServicePort()))
							.setName("http")
							);
					return;
				case GRPC:
					svcSpecBuilder.addPorts(ServicePort.newBuilder()
							.setProtocol("TCP")
							.setPort(e.getServicePort())
							.setTargetPort(IntOrString.newBuilder().setIntVal(e.getServicePort()))
							.setName("grpc")
							);	
					return;
				}
			}
		}
		for(int i=0;i<pu.getChildrenCount();i++)
			addServicePort(pu.getChildren(i), serviceName,svcSpecBuilder);
	}
	
	@Override
	public DeploymentResources createResources(SeldonDeployment mlDep) throws SeldonDeploymentException {
		
		OwnerReference ownerRef = getOwnerReference(mlDep);
		List<Deployment> deployments = new ArrayList<>();
		List<Service> services = new ArrayList<>();
		// for each predictor Create/replace deployment
		final String serviceLabel = mlDep.getSpec().getName();
		Set<String> createdServices = new HashSet<>();
		for(int pbIdx=0;pbIdx<mlDep.getSpec().getPredictorsCount();pbIdx++)
		{
			PredictorSpec p = mlDep.getSpec().getPredictors(pbIdx);

			{//Deployment for engine service orchestrator
				PodTemplateSpec.Builder podSpecBuilder = PodTemplateSpec.newBuilder();
				podSpecBuilder.getSpecBuilder()
		    	.addContainers(createEngineContainer(mlDep,p))
		    	.setTerminationGracePeriodSeconds(20)
		    	.addVolumes(Volume.newBuilder() // Add downwardAPI volume for annotations
		    			.setName(PODINFO_VOLUME_NAME)
		    			.setVolumeSource(VolumeSource.newBuilder().setDownwardAPI(DownwardAPIVolumeSource.newBuilder()
		    			.addItems(DownwardAPIVolumeFile.newBuilder().setPath("annotations")
		    					.setFieldRef(ObjectFieldSelector.newBuilder().setFieldPath("metadata.annotations"))))));
			
				String depName = getSeldonServiceName(mlDep,p,"svc-orch");
				podSpecBuilder.getMetadataBuilder()
					.putAllAnnotations(mlDep.getSpec().getAnnotationsMap()) // Add all spec annotations first
					.putAllAnnotations(p.getAnnotationsMap()) // ...then add those for predictor overwriting any from spec above
			    	.putAnnotations("prometheus.io/path", "/prometheus")
			    	.putAnnotations("prometheus.io/port",""+clusterManagerProperites.getEngineContainerPort())
			    	.putAnnotations("prometheus.io/scrape", "true");

				ObjectMeta.Builder depMetaBuilder = ObjectMeta.newBuilder()
						.setName(depName)
						
						.addOwnerReferences(ownerRef);

				// LABELS - START
				podSpecBuilder.getMetadataBuilder()
					.putLabels(LABEL_SELDON_APP, serviceLabel) 
					.putLabels("app", depName);
				depMetaBuilder
					.putLabels(LABEL_SELDON_APP, serviceLabel)
					.putLabels(Constants.LABEL_SELDON_ID, mlDep.getSpec().getName())
					.putLabels("app", depName)
					.putLabels("version", "v1") // Add default version
					.putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_KEY, SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_VAL);
				// Add all labels from the predictor but don't allow overwriting of key label for deployment selector
				for(Map.Entry<String, String> predictorLabel : p.getLabelsMap().entrySet())
				{
					if (!predictorLabel.getKey().equals(LABEL_SELDON_APP))
					{
						depMetaBuilder.putLabels(predictorLabel.getKey(), predictorLabel.getValue());
						podSpecBuilder.getMetadataBuilder().putLabels(predictorLabel.getKey(), predictorLabel.getValue());
					}
				}
				// LABELS - END
				
				Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
						.setMetadata(depMetaBuilder)
						.setSpec(DeploymentSpec.newBuilder()
						        .setTemplate(podSpecBuilder.build())
						        .setStrategy(DeploymentStrategy.newBuilder().setRollingUpdate(RollingUpdateDeployment.newBuilder().setMaxUnavailable(IntOrString.newBuilder().setType(1).setStrVal("10%"))))
								.setReplicas(p.getReplicas())
								.setSelector(LabelSelector.newBuilder().putMatchLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel))
								)
						.build();
				
				deployments.add(deployment);
			}
			
			
			for(int ptsIdx=0;ptsIdx<p.getComponentSpecsCount();ptsIdx++)
			{
				V1.PodTemplateSpec spec = p.getComponentSpecs(ptsIdx);
				final String depName = getSeldonServiceName(mlDep,p,spec.getSpec().getContainers(0).getName()+"-"+ptsIdx);
				PodTemplateSpec.Builder podSpecBuilder = PodTemplateSpec.newBuilder(spec);
				ObjectMeta.Builder depMetaBuilder = ObjectMeta.newBuilder()
						.setName(depName)
						.addOwnerReferences(ownerRef);
				
				// LABELS - START
				depMetaBuilder.putLabels(Constants.LABEL_SELDON_ID, mlDep.getSpec().getName())
					.putLabels("app", depName)
					.putLabels(SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_KEY, SeldonDeploymentOperatorImpl.LABEL_SELDON_TYPE_VAL);
				podSpecBuilder.getMetadataBuilder().putLabels(Constants.LABEL_SELDON_ID, mlDep.getSpec().getName());
				// Add labels from the predictor for this deployment but not overwriting existing labels
				for(Map.Entry<String, String> predictorLabel : p.getLabelsMap().entrySet())
				{
					if (!depMetaBuilder.containsLabels(predictorLabel.getKey()))
						depMetaBuilder.putLabels(predictorLabel.getKey(), predictorLabel.getValue());
					if (!podSpecBuilder.getMetadataBuilder().containsLabels(predictorLabel.getKey()))
						podSpecBuilder.getMetadataBuilder().putLabels(predictorLabel.getKey(), predictorLabel.getValue());
				}
				// LABELS - END

				podSpecBuilder.getMetadataBuilder()
					.putAllAnnotations(mlDep.getSpec().getAnnotationsMap()) // Add all spec annotations first
					.putAllAnnotations(p.getAnnotationsMap()); // ...then add those for predictor overwriting any from spec above
				
				podSpecBuilder.getSpecBuilder()
		    	.setTerminationGracePeriodSeconds(20)
		    	.addVolumes(Volume.newBuilder() // Add downwardAPI volume for annotations
		    			.setName(PODINFO_VOLUME_NAME)
		    			.setVolumeSource(VolumeSource.newBuilder().setDownwardAPI(DownwardAPIVolumeSource.newBuilder()
		    			.addItems(DownwardAPIVolumeFile.newBuilder().setPath("annotations")
		    					.setFieldRef(ObjectFieldSelector.newBuilder().setFieldPath("metadata.annotations"))))));
				
				LabelSelector.Builder labelSelector = LabelSelector.newBuilder();
				for(V1.Container c : spec.getSpec().getContainersList())
				{
					final String containerServiceKey = getPredictorServiceNameKey(c.getName());
					final String containerServiceValue = getSeldonServiceName(mlDep, p, c.getName());
					
					if (!createdServices.contains(containerServiceValue))
					{
						//Add service
						Service.Builder s = Service.newBuilder()
								.setMetadata(ObjectMeta.newBuilder()
										.setName(containerServiceValue)
										.putLabels(containerServiceKey, containerServiceValue)
										.putLabels("seldon-deployment-id", mlDep.getSpec().getName())
										.addOwnerReferences(ownerRef)
										);
						ServiceSpec.Builder svcSpecBuilder = ServiceSpec.newBuilder();
						addServicePort(p.getGraph(), containerServiceValue, svcSpecBuilder);
						svcSpecBuilder.setType("ClusterIP").putSelector(containerServiceKey,containerServiceValue);

						// LABELS - START
						// Add the service selector label to both deployment, pod and Service Selector
						depMetaBuilder.putLabels(containerServiceKey, containerServiceValue);
						podSpecBuilder.getMetadataBuilder().putLabels(containerServiceKey, containerServiceValue);
						labelSelector.putMatchLabels(containerServiceKey, containerServiceValue);
						// LABELS - END
						s.setSpec(svcSpecBuilder);
						services.add(s.build());
					}
				}
				
				
				
				
				Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
						.setMetadata(depMetaBuilder)
						.setSpec(DeploymentSpec.newBuilder()
						        .setTemplate(podSpecBuilder.build())
						        .setStrategy(DeploymentStrategy.newBuilder().setRollingUpdate(RollingUpdateDeployment.newBuilder().setMaxUnavailable(IntOrString.newBuilder().setType(1).setStrVal("10%"))))
								.setReplicas(p.getReplicas())
								.setSelector(labelSelector.build())
								)
								
						.build();
				
				deployments.add(deployment);
			}
		}
		
		Service s = Service.newBuilder()
					.setMetadata(ObjectMeta.newBuilder()
							.setName(serviceLabel)
							.putLabels(LABEL_SELDON_APP, serviceLabel)
							.putLabels("seldon-deployment-id", mlDep.getSpec().getName())
							.addOwnerReferences(ownerRef)
							.putAnnotations("getambassador.io/config",getAmbassadorAnnotation(mlDep,serviceLabel))
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
		
		services.add(s);
		

		
		// Create service for deployment
		return new DeploymentResources(deployments, services);
	}
	

	public static class DeploymentResources {
		
		List<Deployment> deployments;
		List<Service> services;
		
		public DeploymentResources(List<Deployment> deployments, List<Service> services) {
			super();
			this.deployments = deployments;
			this.services = services;
		}
		

	}


	
}
