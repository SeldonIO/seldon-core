package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.StringJoiner;
import java.util.function.Consumer;

import org.apache.commons.lang3.builder.ReflectionToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.protobuf.InvalidProtocolBufferException;

import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.LocalObjectReference;
import io.fabric8.kubernetes.api.model.Quantity;
import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.ServiceBuilder;
import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.fabric8.kubernetes.api.model.extensions.DeploymentBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.EndpointDef.EndpointType;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.PredictiveUnitSubType;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class DeploymentUtils {

    private final static Logger logger = LoggerFactory.getLogger(DeploymentUtils.class);

    public static class ServiceSelectorDetails {
        public final String appLabelName = "seldon-app";
        public final String appLabelValue;
        public final String trackLabelName = "seldon-track";
        public final String trackLabelValue;
        public final boolean serviceNeeded;

        public ServiceSelectorDetails(String seldonDeploymentId, boolean isCanary) {
            //@formatter:off
            this.appLabelValue = getKubernetesDeploymentId(seldonDeploymentId, false); // Force selector to use the main predictor
            //@formatter:on
            this.trackLabelValue = isCanary ? "canary" : "stable";
            this.serviceNeeded = !isCanary;
        }

        @Override
        public String toString() {
            return ReflectionToStringBuilder.toString(this, ToStringStyle.SHORT_PREFIX_STYLE);
        }
    }

    public static class BuildDeploymentResult {
        public final Deployment deployment;
        public final Optional<Service> service;
        public final PredictorDef resultingPredictorDef;
        public final boolean isCanary;

        public BuildDeploymentResult(Deployment deployment, Optional<Service> service, PredictorDef resultingPredictorDef, boolean isCanary) {
            this.deployment = deployment;
            this.service = service;
            this.resultingPredictorDef = resultingPredictorDef;
            this.isCanary = isCanary;
        }

        @Override
        public String toString() {
            return ReflectionToStringBuilder.toString(this, ToStringStyle.SHORT_PREFIX_STYLE);
        }

    }

    public static List<BuildDeploymentResult> buildDeployments(DeploymentDef deploymentDef, ClusterManagerProperites clusterManagerProperites) {

        final String seldonDeploymentId = deploymentDef.getId();
        List<BuildDeploymentResult> buildDeploymentResults = new ArrayList<>();

        { // Add the main predictor
            PredictorDef mainPredictor = deploymentDef.getPredictor();
            boolean isCanary = false;
            BuildDeploymentResult buildDeploymentResult = buildDeployment(seldonDeploymentId, mainPredictor, isCanary, clusterManagerProperites);
            buildDeploymentResults.add(buildDeploymentResult);
        }

        { // Add the canary predictor if it exists
            if (deploymentDef.hasField(deploymentDef.getDescriptorForType().findFieldByNumber(DeploymentDef.PREDICTOR_CANARY_FIELD_NUMBER))) {
                PredictorDef canaryPredictor = deploymentDef.getPredictorCanary();
                boolean isCanary = true;
                BuildDeploymentResult buildDeploymentResult = buildDeployment(seldonDeploymentId, canaryPredictor, isCanary, clusterManagerProperites);
                buildDeploymentResults.add(buildDeploymentResult);
            }
        }

        return buildDeploymentResults;
    }

    public static BuildDeploymentResult buildDeployment(String seldonDeploymentId, PredictorDef predictorDef, boolean isCanary,
            ClusterManagerProperites clusterManagerProperites) {

        PredictorDef.Builder resultingPredictorDefBuilder = PredictorDef.newBuilder(predictorDef);

        final EndpointType ENGINE_CONTAINER_ENDPOINT_TYPE = EndpointDef.EndpointType.REST;
        final int ENGINE_CONTAINER_PORT = clusterManagerProperites.getEngineContainerPort();
        final String ENGINE_CONTAINER_IMAGE_AND_VERSION = clusterManagerProperites.getEngineContainerImageAndVersion();
        final int PU_CONTAINER_PORT_BASE = clusterManagerProperites.getPuContainerPortBase();

        List<Container> containers = new ArrayList<>();
        List<Service> services = new ArrayList<>();

        final String kubernetesDeploymentId = getKubernetesDeploymentId(seldonDeploymentId, isCanary);

        List<PredictiveUnitDef> predictiveUnits = predictorDef.getPredictiveUnitsList();
        int predictiveUnitIndex = -1;
        for (PredictiveUnitDef predictiveUnitDef : predictiveUnits) {
            predictiveUnitIndex++;

            if (!isContainerRequired(predictiveUnitDef)) {
                logger.debug("IGNORE provision for container of predictiveUnit name[{}] type[{}] subtype[{}]", predictiveUnitDef.getName(),
                        predictiveUnitDef.getType(), predictiveUnitDef.getSubtype());
                continue; // only create container details for predictiveUnit that need it (eg. subtype is external)
            }

            final ClusterResourcesDef clusterResourcesDef = predictiveUnitDef.getClusterResources();

            final int container_port = PU_CONTAINER_PORT_BASE + predictiveUnitIndex;
            final int service_port = container_port;

            final String predictiveUnitParameters = extractPredictiveUnitParametersAsJson(predictiveUnitDef);

            final String image_name_and_version = (clusterResourcesDef.getVersion().length() > 0)
                    ? clusterResourcesDef.getImage() + ":" + clusterResourcesDef.getVersion() : clusterResourcesDef.getImage();

            EnvVar envVar_PREDICTIVE_UNIT_PARAMETERS = new EnvVarBuilder().withName("PREDICTIVE_UNIT_PARAMETERS").withValue(predictiveUnitParameters).build();
            EnvVar envVar_PREDICTIVE_UNIT_SERVICE_PORT = new EnvVarBuilder().withName("PREDICTIVE_UNIT_SERVICE_PORT").withValue(String.valueOf(container_port))
                    .build();

            Map<String, Quantity> resource_requests = new HashMap<>();
            { // Add container resource requests
                if (clusterResourcesDef.hasField(clusterResourcesDef.getDescriptorForType().findFieldByNumber(ClusterResourcesDef.CPU_FIELD_NUMBER))) {
                    resource_requests.put("cpu", new Quantity(clusterResourcesDef.getCpu()));
                }
                if (clusterResourcesDef.hasField(clusterResourcesDef.getDescriptorForType().findFieldByNumber(ClusterResourcesDef.MEMORY_FIELD_NUMBER))) {
                    resource_requests.put("memory", new Quantity(clusterResourcesDef.getMemory()));
                }
            }

            //@formatter:off
            Container c = new ContainerBuilder()
                    .withName("seldon-container-pu-"+String.valueOf(predictiveUnitIndex)).withImage(image_name_and_version)
                    .withEnv(envVar_PREDICTIVE_UNIT_PARAMETERS, envVar_PREDICTIVE_UNIT_SERVICE_PORT)
                    .addNewPort().withContainerPort(container_port).endPort()
                    .withNewResources()
                        .addToRequests(resource_requests)
                    .endResources()
                    .build();
            
            containers.add(c);
            //@formatter:on
            logger.debug("ADDING provision for container of predictiveUnit name[{}] type[{}] subtype[{}] image[{}]", predictiveUnitDef.getName(),
                    predictiveUnitDef.getType(), predictiveUnitDef.getSubtype(), image_name_and_version);

            { // update the resulting predictorDef with the host/port details for this predictive
                resultingPredictorDefBuilder.getPredictiveUnitsBuilder(predictiveUnitIndex).getEndpointBuilder().setServiceHost("localhost");
                resultingPredictorDefBuilder.getPredictiveUnitsBuilder(predictiveUnitIndex).getEndpointBuilder().setServicePort(service_port);
            }
        }

        final int replica_number = predictorDef.getReplicas();

        List<LocalObjectReference> imagePullSecrets = new ArrayList<>();
        { // add any image pull secrets
            Consumer<String> p = (x) -> {
                LocalObjectReference imagePullSecretObject = new LocalObjectReference(x);
                imagePullSecrets.add(imagePullSecretObject);
            };
            predictorDef.getImagePullSecretsList().forEach(p);
        }

        final int engine_container_port = ENGINE_CONTAINER_PORT;
        final int engine_service_port = engine_container_port;
        { // add container for engine
            final String image_name_and_version = ENGINE_CONTAINER_IMAGE_AND_VERSION;

            String enginePredictorJson = getEnginePredictorEnvVarJson(resultingPredictorDefBuilder.build());
            EnvVar envVar_ENGINE_PREDICTOR = new EnvVarBuilder().withName("ENGINE_PREDICTOR").withValue(enginePredictorJson).build();
            EnvVar envVar_ENGINE_SERVER_PORT = new EnvVarBuilder().withName("ENGINE_SERVER_PORT").withValue(String.valueOf(engine_container_port)).build();

            //@formatter:off
            Container c = new ContainerBuilder()
                    .withName("seldon-container-engine").withImage(image_name_and_version)
                    .withEnv(envVar_ENGINE_PREDICTOR, envVar_ENGINE_SERVER_PORT)
                    .addNewPort().withContainerPort(engine_container_port).endPort()
                    .build();
            
            containers.add(c);
            //@formatter:on
            logger.debug("ADDING provision for container of seldon engine");
        }

        ServiceSelectorDetails serviceSelectorDetails = new ServiceSelectorDetails(seldonDeploymentId, isCanary);
        Service service = null;
        if (serviceSelectorDetails.serviceNeeded) { // build service for this predictor
            final String deploymentName = kubernetesDeploymentId;
            String serviceName = deploymentName;

            String selectorName = serviceSelectorDetails.appLabelName;
            String selectorValue = serviceSelectorDetails.appLabelValue;

            int port = engine_service_port;
            int targetPort = engine_container_port;

            //@formatter:off
            service = new ServiceBuilder()
                    .withNewMetadata()
                        .withName(serviceName)
                        .addToLabels("seldon-deployment-id", seldonDeploymentId)
                    .endMetadata()
                    .withNewSpec()
                        .addNewPort()
                            .withProtocol("TCP")
                            .withPort(port)
                            .withNewTargetPort(targetPort)
                        .endPort()
                        .addToSelector(selectorName, selectorValue)
                        .withType("ClusterIP")
                    .endSpec()
                    .build();
            //@formatter:on
            services.add(service);

            //@formatter:off
            resultingPredictorDefBuilder.setEndpoint(EndpointDef.newBuilder()
                    .setServiceHost(serviceName)
                    .setServicePort(port)
                    .setType(ENGINE_CONTAINER_ENDPOINT_TYPE)
                    .build());
            //@formatter:on
        }

        //@formatter:off
        Deployment deployment = new DeploymentBuilder()
            .withNewMetadata().withName(kubernetesDeploymentId).addToLabels("seldon-deployment-id", seldonDeploymentId).endMetadata()
            .withNewSpec().withReplicas(replica_number)
                .withNewTemplate()
                    .withNewMetadata()
                        .addToLabels(serviceSelectorDetails.appLabelName, serviceSelectorDetails.appLabelValue)
                        .addToLabels(serviceSelectorDetails.trackLabelName, serviceSelectorDetails.trackLabelValue)
                    .endMetadata()
                    .withNewSpec()
                        .addAllToContainers(containers)
                        .addAllToImagePullSecrets(imagePullSecrets)
                    .endSpec()
                .endTemplate()
            .endSpec().build();
        //@formatter:on

        BuildDeploymentResult buildDeploymentResult = new BuildDeploymentResult(deployment, Optional.ofNullable(service), resultingPredictorDefBuilder.build(),
                isCanary);
        return buildDeploymentResult;
    }

    public static void createDeployment(KubernetesClient kubernetesClient, String namespace_name, BuildDeploymentResult buildDeploymentResult) {
        Deployment deployment = kubernetesClient.extensions().deployments().inNamespace(namespace_name).createOrReplace(buildDeploymentResult.deployment);
        String deploymentName = (deployment != null) ? deployment.getMetadata().getName() : "null";
        logger.debug(String.format("Created kubernetes delployment [%s]", deploymentName));
        if ((deployment != null) && (buildDeploymentResult.service.isPresent())) {
            Service service = kubernetesClient.services().inNamespace(namespace_name).createOrReplace(buildDeploymentResult.service.get());
            String serviceName = (service != null) ? service.getMetadata().getName() : "null";
            logger.debug(String.format("Created kubernetes service [%s]", serviceName));
        }
    }

    public static void deleteDeployment(KubernetesClient kubernetesClient, String namespace_name, DeploymentDef deploymentDef) {
        final String seldonDeploymentId = deploymentDef.getId();

        { // delete the services for this seldon deployment

            io.fabric8.kubernetes.api.model.ServiceList svcList = kubernetesClient.services().inNamespace(namespace_name)
                    .withLabel("seldon-deployment-id", seldonDeploymentId).list();
            for (io.fabric8.kubernetes.api.model.Service service : svcList.getItems()) {
                kubernetesClient.resource(service).inNamespace(namespace_name).delete();
                String rsmsg = (service != null) ? service.getMetadata().getName() : null;
                logger.debug(String.format("Deleted kubernetes service [%s]", rsmsg));
            }
        }

        //@formatter:off
        deleteDeployemntResources(kubernetesClient, namespace_name, seldonDeploymentId, false); // for the main predictor
        deleteDeployemntResources(kubernetesClient, namespace_name, seldonDeploymentId, true); // for the canary
        //@formatter:on
    }

    public static void deleteDeployemntResources(KubernetesClient kubernetesClient, String namespace_name, String seldonDeploymentId, boolean isCanary) {
        final String kubernetesDeploymentId = getKubernetesDeploymentId(seldonDeploymentId, isCanary);
        final String deploymentName = kubernetesDeploymentId;
        boolean wasDeleted = kubernetesClient.extensions().deployments().inNamespace(namespace_name).withName(deploymentName).delete();
        if (wasDeleted) {
            logger.debug(String.format("Deleted kubernetes delployment [%s]", deploymentName));
        }
        io.fabric8.kubernetes.api.model.extensions.ReplicaSetList rslist = kubernetesClient.extensions().replicaSets().inNamespace(namespace_name)
                .withLabel("seldon-app", deploymentName).list();
        for (io.fabric8.kubernetes.api.model.extensions.ReplicaSet rs : rslist.getItems()) {
            kubernetesClient.resource(rs).inNamespace(namespace_name).delete();
            String rsmsg = (rs != null) ? rs.getMetadata().getName() : null;
            logger.debug(String.format("Deleted kubernetes replicaSet [%s]", rsmsg));
        }

    }

    public static DeploymentDef getDeployments(KubernetesClient kubernetesClient, String namespace_name, DeploymentDef deploymentDef) {

        DeploymentDef.Builder resultingDeploymentDefBuilder = DeploymentDef.newBuilder(deploymentDef);

        final String seldonDeploymentId = deploymentDef.getId();

        Consumer<Boolean> updateReplicasReady = (isCanary) -> {
            final String kubernetesDeploymentId = getKubernetesDeploymentId(seldonDeploymentId, isCanary);
            int replicasReady = 0;
            Deployment deployment = kubernetesClient.extensions().deployments().inNamespace(namespace_name).withName(kubernetesDeploymentId).get();
            if (deployment != null) {
                Integer readyReplicasValue = (Integer) deployment.getStatus().getAdditionalProperties().get("readyReplicas");
                if (readyReplicasValue != null) {
                    replicasReady = readyReplicasValue;
                }
            }

            PredictorDef.Builder predictorDefBuilder = (isCanary) ? resultingDeploymentDefBuilder.getPredictorCanaryBuilder()
                    : resultingDeploymentDefBuilder.getPredictorBuilder();
            predictorDefBuilder.setReplicasReady(replicasReady);
        };

        updateReplicasReady.accept(false); // for main predictor
        if (deploymentDef.hasField(deploymentDef.getDescriptorForType().findFieldByNumber(DeploymentDef.PREDICTOR_CANARY_FIELD_NUMBER))) {
            updateReplicasReady.accept(true); // for canary predictor
        }

        return resultingDeploymentDefBuilder.build();
    }

    private static String extractPredictiveUnitParametersAsJson(PredictiveUnitDef predictiveUnitDef) {
        StringJoiner sj = new StringJoiner(",", "[", "]");
        List<ParamDef> parameters = predictiveUnitDef.getParametersList();
        for (ParamDef parameter : parameters) {
            try {
                String j = ProtoBufUtils.toJson(parameter, true);
                sj.add(j);
            } catch (InvalidProtocolBufferException e) {
                throw new RuntimeException(e);
            }
        }
        return sj.toString();
    }

    private static boolean isContainerRequired(PredictiveUnitDef predictiveUnitDef) {
        // Predictive units that have the subtype "microservice" are the only ones that need a container to be provisioned
        return predictiveUnitDef.getSubtype().equals(PredictiveUnitSubType.MICROSERVICE);
    }

    private static String getKubernetesDeploymentId(String seldonDeploymentId, boolean isCanary) {
        return "sd-" + seldonDeploymentId + "-" + ((isCanary) ? "c" : "p");
    }

    public static String getEnginePredictorEnvVarJson(PredictorDef predictorDef) {
        String retVal;
        try {
            retVal = ProtoBufUtils.toJson(predictorDef, true);
        } catch (InvalidProtocolBufferException e) {
            retVal = e.getMessage();
        }

        retVal = new String(Base64.getEncoder().encode(retVal.getBytes()));

        return retVal;
    }
}
