package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.StringJoiner;

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
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class DeploymentUtils {

    private final static Logger logger = LoggerFactory.getLogger(DeploymentUtils.class);

    public static class BuildDeploymentResult {
        public final Deployment deployment;
        public final List<Service> services;
        public final Map<String, EndpointDef> endpointsByPredictiveUnitId;

        public BuildDeploymentResult(Deployment deployment, List<Service> services, Map<String, EndpointDef> endpointsByPredictiveUnitId) {
            this.deployment = deployment;
            this.services = services;
            this.endpointsByPredictiveUnitId = endpointsByPredictiveUnitId;
        }
    }

    public static BuildDeploymentResult buildDeployment(DeploymentDef deploymentDef) {

        final int CONTAINER_PORT_BASE = 5000;

        final String seldonDeploymentId = deploymentDef.getId();
        PredictorDef predictorDef = deploymentDef.getPredictor();

        List<Container> containers = new ArrayList<>();
        List<Service> services = new ArrayList<>();
        Map<String, EndpointDef> endpointsByPredictiveUnitId = new HashMap<>();

        final String kubernetesDeploymentId = "sd-" + seldonDeploymentId;

        List<PredictiveUnitDef> predictiveUnits = predictorDef.getPredictiveUnitsList();
        int predictiveUnitIndex = 0;
        for (PredictiveUnitDef predictiveUnitDef : predictiveUnits) {

            final ClusterResourcesDef clusterResourcesDef = predictiveUnitDef.getClusterResources();
            if (!hasDeployableImage(clusterResourcesDef)) {
                break; // only create container details for predictiveUnit that has an image
            }

            final int container_port = CONTAINER_PORT_BASE + predictiveUnitIndex;
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
                    .withName("seldon-container-"+String.valueOf(predictiveUnitIndex)).withImage(image_name_and_version)
                    .withEnv(envVar_PREDICTIVE_UNIT_PARAMETERS, envVar_PREDICTIVE_UNIT_SERVICE_PORT)
                    .addNewPort().withContainerPort(container_port).endPort()
                    .withNewResources()
                        .addToRequests(resource_requests)
                    .endResources()
                    .build();
            
            containers.add(c);
            //@formatter:on

            { // build service for this predictiveUnit
                final String deploymentName = kubernetesDeploymentId;
                String serviceName = deploymentName + "-" + predictiveUnitDef.getId();

                String selectorName = "seldon-app";
                String selectorValue = deploymentName;

                int port = service_port;
                int targetPort = container_port;

                //@formatter:off
                Service service = new ServiceBuilder()
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

                // build an EndpointDef with service details
                //@formatter:off
                EndpointDef endpointDef = EndpointDef.newBuilder()
                        .setServiceHost(serviceName)
                        .setServicePort(port)
                        .build();
                //@formatter:on
                endpointsByPredictiveUnitId.put(predictiveUnitDef.getId(), endpointDef);
            }

            predictiveUnitIndex++;
        }

        final int replica_number = deploymentDef.getPredictor().getReplicas();
        final String imagePullSecret = ""; // clusterResourcesDef.getImagePullSecret(); TODO!!
        List<LocalObjectReference> imagePullSecrets = new ArrayList<>();
        if (imagePullSecret.length() > 0) {
            LocalObjectReference imagePullSecretObject = new LocalObjectReference(imagePullSecret);
            imagePullSecrets.add(imagePullSecretObject);
        }

        //@formatter:off
        Deployment deployment = new DeploymentBuilder()
            .withNewMetadata().withName(kubernetesDeploymentId).addToLabels("seldon-deployment-id", seldonDeploymentId).endMetadata()
            .withNewSpec().withReplicas(replica_number)
                .withNewTemplate()
                    .withNewMetadata().addToLabels("seldon-app", kubernetesDeploymentId).endMetadata()
                    .withNewSpec()
                        .addAllToContainers(containers)
                        .addAllToImagePullSecrets(imagePullSecrets)
                    .endSpec()
                .endTemplate()
            .endSpec().build();
        //@formatter:on

        BuildDeploymentResult buildDeploymentResult = new BuildDeploymentResult(deployment, services, endpointsByPredictiveUnitId);
        return buildDeploymentResult;
    }

    public static void createDeployment(KubernetesClient kubernetesClient, String namespace_name, BuildDeploymentResult buildDeploymentResult) {
        Deployment deployment = kubernetesClient.extensions().deployments().inNamespace(namespace_name).create(buildDeploymentResult.deployment);
        String deploymentName = (deployment != null) ? deployment.getMetadata().getName() : "null";
        logger.debug(String.format("Created kubernetes delployment [%s]", deploymentName));
        if (deployment != null) {
            for (Service service : buildDeploymentResult.services) {
                service = kubernetesClient.services().inNamespace(namespace_name).create(service);
                String serviceName = (service != null) ? service.getMetadata().getName() : "null";
                logger.debug(String.format("Created kubernetes service [%s]", serviceName));
            }
        }
    }

    public static void deleteDeployment(KubernetesClient kubernetesClient, String namespace_name, DeploymentDef deploymentDef) {
        final String seldonDeploymentId = deploymentDef.getId();
        final String deploymentName = "sd-" + seldonDeploymentId;

        { // delete the services for this seldon deployment

            io.fabric8.kubernetes.api.model.ServiceList svcList = kubernetesClient.services().inNamespace(namespace_name)
                    .withLabel("seldon-deployment-id", seldonDeploymentId).list();
            for (io.fabric8.kubernetes.api.model.Service service : svcList.getItems()) {
                kubernetesClient.resource(service).inNamespace(namespace_name).delete();
                String rsmsg = (service != null) ? service.getMetadata().getName() : null;
                logger.debug(String.format("Deleted kubernetes service [%s]", rsmsg));
            }
        }
        { // delete the kubernetes deployemnt for this seldon deployment
            kubernetesClient.extensions().deployments().inNamespace(namespace_name).withName(deploymentName).delete();
            logger.debug(String.format("Deleted kubernetes delployment [%s]", deploymentName));
            io.fabric8.kubernetes.api.model.extensions.ReplicaSetList rslist = kubernetesClient.extensions().replicaSets().inNamespace(namespace_name)
                    .withLabel("seldon-app", deploymentName).list();
            for (io.fabric8.kubernetes.api.model.extensions.ReplicaSet rs : rslist.getItems()) {
                kubernetesClient.resource(rs).inNamespace(namespace_name).delete();
                String rsmsg = (rs != null) ? rs.getMetadata().getName() : null;
                logger.debug(String.format("Deleted kubernetes replicaSet [%s]", rsmsg));
            }
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

}
