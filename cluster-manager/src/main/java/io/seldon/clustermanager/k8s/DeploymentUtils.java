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
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.EndpointDef.EndpointType;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class DeploymentUtils {

    private final static Logger logger = LoggerFactory.getLogger(DeploymentUtils.class);

    public static class ServiceSelectorDetails {
        public final String labelName = "seldon-app";
        public final String labelValue;
        public final boolean serviceNeeded;

        public ServiceSelectorDetails(String seldonDeploymentId, boolean isCanary) {
            //@formatter:off
            this.labelValue = getKubernetesDeploymentId(seldonDeploymentId, false); // Force selector to use the main predictor
            //@formatter:on
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

    public static List<BuildDeploymentResult> buildDeployments(DeploymentDef deploymentDef) {

        final String seldonDeploymentId = deploymentDef.getId();
        List<BuildDeploymentResult> buildDeploymentResults = new ArrayList<>();

        { // Add the main predictor
            PredictorDef mainPredictor = deploymentDef.getPredictor();
            boolean isCanary = false;
            BuildDeploymentResult buildDeploymentResult = buildDeployment(seldonDeploymentId, mainPredictor, isCanary);
            buildDeploymentResults.add(buildDeploymentResult);
        }

        { // Add the canary predictor if it exists
            if (deploymentDef.hasField(deploymentDef.getDescriptorForType().findFieldByNumber(DeploymentDef.PREDICTOR_CANARY_FIELD_NUMBER))) {
                PredictorDef canaryPredictor = deploymentDef.getPredictorCanary();
                boolean isCanary = true;
                BuildDeploymentResult buildDeploymentResult = buildDeployment(seldonDeploymentId, canaryPredictor, isCanary);
                buildDeploymentResults.add(buildDeploymentResult);
            }
        }

        return buildDeploymentResults;
    }

    public static BuildDeploymentResult buildDeployment(String seldonDeploymentId, PredictorDef predictorDef, boolean isCanary) {

        PredictorDef.Builder resultingPredictorDefBuilder = PredictorDef.newBuilder(predictorDef);

        final int ENGINE_CONTAINER_PORT = 8000;
        final String ENGINE_CONTAINER_IMAGE_AND_VERSION = "gsunner/putest";
        final EndpointType ENGINE_CONTAINER_ENDPOINT_TYPE = EndpointDef.EndpointType.REST;
        final int PU_CONTAINER_PORT_BASE = 9000;

        List<Container> containers = new ArrayList<>();
        List<Service> services = new ArrayList<>();

        final String kubernetesDeploymentId = getKubernetesDeploymentId(seldonDeploymentId, isCanary);

        List<PredictiveUnitDef> predictiveUnits = predictorDef.getPredictiveUnitsList();
        int predictiveUnitIndex = 0;
        for (PredictiveUnitDef predictiveUnitDef : predictiveUnits) {

            final ClusterResourcesDef clusterResourcesDef = predictiveUnitDef.getClusterResources();
            if (!hasDeployableImage(clusterResourcesDef)) {
                break; // only create container details for predictiveUnit that has an image
            }

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

            { // update the resulting predictorDef with the host/port details for this predictive
                resultingPredictorDefBuilder.getPredictiveUnitsBuilder(predictiveUnitIndex).getEndpointBuilder().setServiceHost("localhost");
                resultingPredictorDefBuilder.getPredictiveUnitsBuilder(predictiveUnitIndex).getEndpointBuilder().setServicePort(service_port);
            }

            predictiveUnitIndex++;
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

            //@formatter:off
            Container c = new ContainerBuilder()
                    .withName("seldon-container-engine").withImage(image_name_and_version)
                    .withEnv(envVar_ENGINE_PREDICTOR)
                    .addNewPort().withContainerPort(engine_container_port).endPort()
                    .build();
            
            containers.add(c);
            //@formatter:on

        }

        ServiceSelectorDetails serviceSelectorDetails = new ServiceSelectorDetails(seldonDeploymentId, isCanary);
        Service service = null;
        if (serviceSelectorDetails.serviceNeeded) { // build service for this predictor
            final String deploymentName = kubernetesDeploymentId;
            String serviceName = deploymentName;

            String selectorName = serviceSelectorDetails.labelName;
            String selectorValue = serviceSelectorDetails.labelValue;

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
                    .withNewMetadata().addToLabels(serviceSelectorDetails.labelName, serviceSelectorDetails.labelValue).endMetadata()
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

    private static boolean hasDeployableImage(ClusterResourcesDef clusterResourcesDef) {
        return (clusterResourcesDef.getImage().length() > 0);
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
