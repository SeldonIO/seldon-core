package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.Base64;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import com.google.protobuf.InvalidProtocolBufferException;

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
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.MLDeployment;
import io.seldon.protos.DeploymentProtos.PredictorSpec;

@Component
public class MLDeploymentOperatorImpl implements MLDeploymentOperator {

	private final static Logger logger = LoggerFactory.getLogger(MLDeploymentOperatorImpl.class);
	private final ClusterManagerProperites clusterManagerProperites;
	public static final String LABEL_SELDON_APP = "seldon-app";
	@Autowired
	public MLDeploymentOperatorImpl(ClusterManagerProperites clusterManagerProperites) {
		super();
		this.clusterManagerProperites = clusterManagerProperites;
	}

	/*
	@Autowired
	public void setClusterManagerProperites(ClusterManagerProperites clusterManagerProperites) {
		logger.info(String.format("injecting %s", clusterManagerProperites.toString()));
		this.clusterManagerProperites = clusterManagerProperites;
	}
	*/
	 
	public static String getEnginePredictorEnvVarJson(PredictorSpec predictorDef) {
		String retVal;
		try {
			retVal = ProtoBufUtils.toJson(predictorDef, true,false);
		} catch (InvalidProtocolBufferException e) {
			retVal = e.getMessage();
		}

		retVal = new String(Base64.getEncoder().encode(retVal.getBytes()));
		 
		return retVal;
	}
	
	private V1.Container createEngineContainer(PredictorSpec predictorDef)
	{
		V1.Container.Builder cBuilder = V1.Container.newBuilder();
		cBuilder
			.setName("seldon-container-engine")
			.setImage(clusterManagerProperites.getEngineContainerImageAndVersion())
			.addEnv(EnvVar.newBuilder().setName("ENGINE_PREDICTOR").setValue(getEnginePredictorEnvVarJson(predictorDef)))
			.addEnv(EnvVar.newBuilder().setName("ENGINE_SERVER_PORT").setValue(""+clusterManagerProperites.getEngineContainerPort()))
			.addPorts(V1.ContainerPort.newBuilder().setContainerPort(clusterManagerProperites.getEngineContainerPort()))
			.addPorts(V1.ContainerPort.newBuilder().setContainerPort(8082).setName("admin"))
			.setReadinessProbe(Probe.newBuilder().setHandler(Handler.newBuilder()
					.setHttpGet(HTTPGetAction.newBuilder().setPort(IntOrString.newBuilder().setType(1).setStrVal("admin")).setPath("/ready")))
					.setInitialDelaySeconds(5)
					.setPeriodSeconds(5)
					.setFailureThreshold(1)
					.setSuccessThreshold(1)
					.setTimeoutSeconds(2)
					)
			.setLivenessProbe(Probe.newBuilder().setHandler(Handler.newBuilder()
					.setHttpGet(HTTPGetAction.newBuilder().setPort(IntOrString.newBuilder().setType(1).setStrVal("admin")).setPath("/ready")))
					.setInitialDelaySeconds(5)
					.setPeriodSeconds(5)
					.setFailureThreshold(1)
					.setSuccessThreshold(1)
					.setTimeoutSeconds(2)
					)
			.setLifecycle(Lifecycle.newBuilder().setPreStop(Handler.newBuilder().setExec(
					ExecAction.newBuilder()
						.addCommand("/bin/bash")
						.addCommand("-c")
						.addCommand("curl 127.0.0.1:"+clusterManagerProperites.getEngineContainerPort()+"/pause && /bin/sleep 20"))));

			
			
		return cBuilder.build();
	}
	
	/*
	private V1Container createEngineContainerOld(PredictorDef predictorDef)
	{
		//@formatter:off
		V1Container c = new V1Container()
							.name("seldon-container-engine")
							.image(clusterManagerProperites.getEngineContainerImageAndVersion())
							.addEnvItem(new V1EnvVar().name("ENGINE_PREDICTOR").value(getEnginePredictorEnvVarJson(predictorDef)))
							.addEnvItem(new V1EnvVar().name("ENGINE_SERVER_PORT").value(""+clusterManagerProperites.getEngineContainerPort()))
							.addPortsItem(new V1ContainerPort().containerPort(clusterManagerProperites.getEngineContainerPort()))
							.addPortsItem(new V1ContainerPort().containerPort(8082).name("admin"))
							.readinessProbe(new V1Probe()
									.httpGet(new V1HTTPGetAction().port(new io.kubernetes.client.custom.IntOrString("admin")).path("/ready"))
									.initialDelaySeconds(5)
									.periodSeconds(5)
									.failureThreshold(1)
									.successThreshold(1)
									.timeoutSeconds(2)
									)
							.livenessProbe(new V1Probe()
									.httpGet(new V1HTTPGetAction().port(new io.kubernetes.client.custom.IntOrString("admin")).path("/ping"))
									.initialDelaySeconds(5)
									.periodSeconds(5)
									.failureThreshold(1)
									.successThreshold(1)
									.timeoutSeconds(2)									
									)
							.lifecycle(new V1Lifecycle()
									.preStop(new V1Handler()
											.exec(new V1ExecAction().addCommandItem("/bin/bash").addCommandItem("-c").addCommandItem("curl 127.0.0.1:"+clusterManagerProperites.getEngineContainerPort()+"/pause && /bin/sleep 20"))
											)
									);
		//@formatter:on
		return c;
	}
	*/
	
	private Set<String> getEnvNamesProto(List<EnvVar> envs)
	{
		Set<String> s = new HashSet<>();
		for(EnvVar e : envs)
			s.add(e.getName());
		return s;
	}
	
	private V1.Container updateContainer(V1.Container c,int idx)
	{
		V1.Container.Builder c2Builder = V1.Container.newBuilder(c);
		int containerPort = 9000;
		if (c.getPortsCount() == 0)
		{
			c2Builder.addPorts(ContainerPort.newBuilder().setName("http").setContainerPort(clusterManagerProperites.getPuContainerPortBase() + idx));
			containerPort = clusterManagerProperites.getPuContainerPortBase() + idx;
		}
		else
			containerPort = c.getPorts(0).getContainerPort();
		
		final String ENV_PREDICTIVE_UNIT_SERVICE_PORT ="PREDICTIVE_UNIT_SERVICE_PORT";
		Set<String> envNames = this.getEnvNamesProto(c.getEnvList());
		if (!envNames.contains(ENV_PREDICTIVE_UNIT_SERVICE_PORT))
			c2Builder.addEnv(EnvVar.newBuilder().setName(ENV_PREDICTIVE_UNIT_SERVICE_PORT).setValue(""+containerPort));
				
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
						ExecAction.newBuilder().addCommand("/bin/bash").addCommand("-c").addCommand("/bin/sleep 20"))));
			}
		}

		
		return c2Builder.build();
	}
	
	@Override
	public MLDeployment defaulting(MLDeployment mlDep) {
		MLDeployment.Builder mlBuilder = MLDeployment.newBuilder(mlDep);
		int idx = 0;
		String serviceName = getKubernetesMLDeploymentId(mlDep.getSpec().getName(), false);
		for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
		{
			ObjectMeta.Builder metaBuilder = ObjectMeta.newBuilder(p.getComponentSpec().getMetadata())
				.putLabels(LABEL_SELDON_APP, serviceName);
			mlBuilder.getSpecBuilder().getPredictorsBuilder(idx).getComponentSpecBuilder().setMetadata(metaBuilder);
			int cIdx = 0;
			mlBuilder.getSpecBuilder().getPredictorsBuilder(idx).getComponentSpecBuilder().getSpecBuilder().clearContainers();
			for(V1.Container c : p.getComponentSpec().getSpec().getContainersList())
			{
				V1.Container c2 = this.updateContainer(c, cIdx);
				mlBuilder.getSpecBuilder().getPredictorsBuilder(idx).getComponentSpecBuilder().getSpecBuilder().addContainers(cIdx, c2);				
				cIdx++;
			}
			idx++;
		}
		return mlBuilder.build();
	}

	@Override
	public void validate(MLDeployment mlDep) throws MLDeploymentException {
		// TODO Auto-generated method stub
		
	}
	
	private static String getKubernetesMLDeploymentId(String deploymentName, boolean isCanary) {
		return "sd-" + deploymentName + "-" + ((isCanary) ? "c" : "p");
	}

	private static String getKubernetesDeploymentId(String deploymentName,String predictorName, boolean isCanary) {
		return "sd-" + deploymentName + "-" + predictorName + "-" + ((isCanary) ? "c" : "p");
	}
	
	private V1OwnerReference getOwnerReferenceOld(MLDeployment mlDep)
	{
		return new V1OwnerReference()
				.apiVersion(mlDep.getApiVersion())
				.kind(mlDep.getKind())
				.controller(true)
				.name(mlDep.getMetadata().getName())
				.uid(mlDep.getMetadata().getUid());
	}
	private OwnerReference getOwnerReference(MLDeployment mlDep)
	{
		return OwnerReference.newBuilder()
			.setApiVersion(mlDep.getApiVersion())
			.setKind(mlDep.getKind())
			.setController(true)
			.setName(mlDep.getMetadata().getName())
			.setUid(mlDep.getMetadata().getUid()).build();
	}
	 
	@Override
	public DeploymentResources createResources(MLDeployment mlDep) throws MLDeploymentException {
		
		OwnerReference ownerRef = getOwnerReference(mlDep);
		List<Deployment> deployments = new ArrayList<>();
		// for each predictor Create/replace deployment
		String serviceLabel = getKubernetesMLDeploymentId(mlDep.getSpec().getName(), false);
		for(PredictorSpec p : mlDep.getSpec().getPredictorsList())
		{
			String depName = getKubernetesDeploymentId(mlDep.getSpec().getName(),p.getName(), p.getType().equals(PredictorSpec.PredictorType.CANARY));
			PodTemplateSpec.Builder podSpecBuilder = PodTemplateSpec.newBuilder(p.getComponentSpec());
			podSpecBuilder.getSpecBuilder().addContainers(createEngineContainer(p));
			Deployment deployment = V1beta1Extensions.Deployment.newBuilder()
					.setMetadata(ObjectMeta.newBuilder()
							.setName(depName)
							.putLabels(MLDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel)
							.putLabels("seldon-deployment-id", mlDep.getSpec().getName())
							.putLabels("app", depName)
							.putLabels("version", "v1") //FIXME
							.putLabels("seldon-type", "mldeployment")
							.addOwnerReferences(ownerRef)
							.putAnnotations("prometheus.io/path", "/prometheus")
							.putAnnotations("prometheus.io/port",""+clusterManagerProperites.getEngineContainerPort())
							.putAnnotations("prometheus.io/scrape", "true")
							)
					.setSpec(DeploymentSpec.newBuilder().setTemplate(podSpecBuilder.build())
							.setReplicas(p.getReplicas()))
					.build();
			
			deployments.add(deployment);
		}
		
		Service s = Service.newBuilder()
					.setMetadata(ObjectMeta.newBuilder()
							.setName(mlDep.getSpec().getName())
							.putLabels(MLDeploymentOperatorImpl.LABEL_SELDON_APP, serviceLabel)
							.putLabels("seldon-deployment-id", mlDep.getSpec().getName())
							)
					.setSpec(ServiceSpec.newBuilder()
							.addPorts(ServicePort.newBuilder()
									.setProtocol("TCP")
									.setPort(clusterManagerProperites.getEngineContainerPort())
									.setTargetPort(IntOrString.newBuilder().setIntVal(clusterManagerProperites.getEngineContainerPort()))
									.setName("http")
									)
							.setType("ClusterIP")
							.putSelector(MLDeploymentOperatorImpl.LABEL_SELDON_APP,serviceLabel)
							)
				.build();
		
		// Create service for deployment
		return new DeploymentResources(deployments, s);
	}
	
	/*
	public DeploymentResources createResourcesOld(MLDeployment mlDep) throws MLDeploymentException {
		
		try
		{
			V1OwnerReference ownerRef = getOwnerReferenceOld(mlDep);
			List<ExtensionsV1beta1Deployment> deployments = new ArrayList<>();
			// for each predictor Create/replace deployment
			for(PredictorDef p : mlDep.getSpec().getPredictorsList())
			{
				V1PodTemplateSpec podTemplate = MLDeploymentUtils.convertProtoToModel(p.getComponentSpec());
				V1Container engineContainer = createEngineContainerOld(p);
				podTemplate.getSpec().addContainersItem(engineContainer);
				
				String depName = getKubernetesDeploymentId(mlDep.getSpec().getName(),p.getName(), p.getType().equals(PredictorDef.PredictorType.CANARY));
				String serviceLabel = getKubernetesDeploymentId(mlDep.getSpec().getName(),p.getName(), false);
				ExtensionsV1beta1DeploymentSpec depSpec = new ExtensionsV1beta1DeploymentSpec()
						.template(podTemplate)
						.replicas(p.getReplicas());
						//.selector(new V1LabelSelector().putMatchLabelsItem("seldon-app", depName));
				 ExtensionsV1beta1Deployment dep = new ExtensionsV1beta1Deployment().apiVersion("extensions/v1beta1").kind("Deployment")
						 	.metadata(new V1ObjectMeta()
						 			.name(depName)
						 			.putLabelsItem(LABEL_SELDON_APP, serviceLabel)
						 			.putLabelsItem("seldon-deployment-id", mlDep.getSpec().getName())
						 			.putLabelsItem("app", depName)
						 			.putLabelsItem("version", "v1")  //FIXME
						 			.putLabelsItem("seldon-type", "mldeployment")
						 			.addOwnerReferencesItem(ownerRef)
						 			)
						 	.spec(depSpec);
				 logger.info(dep.toString());
				 deployments.add(dep);
			}
			// Create service for deployment
			return new DeploymentResources(deployments, null);
			
		} catch (InvalidProtocolBufferException e) {
			logger.error("Failed to reconcile ",e);
			throw new MLDeploymentException(e.getMessage());
		}
	}
*/
	
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
