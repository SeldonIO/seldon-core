package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.StringJoiner;

import com.google.protobuf.InvalidProtocolBufferException;

import io.fabric8.kubernetes.api.model.Container;
import io.fabric8.kubernetes.api.model.ContainerBuilder;
import io.fabric8.kubernetes.api.model.EnvVar;
import io.fabric8.kubernetes.api.model.EnvVarBuilder;
import io.fabric8.kubernetes.api.model.LocalObjectReference;
import io.fabric8.kubernetes.api.model.Quantity;
import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.fabric8.kubernetes.api.model.extensions.DeploymentBuilder;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef.ParamDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;

public class DeploymentUtils {

    public static Deployment buildKubernetesDeployment(DeploymentDef deploymentDef) {

        final String seldonDeploymentId = deploymentDef.getId();
        PredictorDef predictorDef = deploymentDef.getPredictor();

        List<Container> containers = new ArrayList<>();

        List<PredictiveUnitDef> predictiveUnits = predictorDef.getPredictiveUnitsList();
        int predictiveUnitIndex = 0;
        for (PredictiveUnitDef predictiveUnitDef : predictiveUnits) {
            final ClusterResourcesDef clusterResourcesDef = predictiveUnitDef.getClusterResources();
            EndpointDef endpointDef = predictiveUnitDef.getEndpoint();
            final String predictiveUnitParameters = extractPredictiveUnitParametersAsJson(predictiveUnitDef);

            final int container_port = endpointDef.getContainerPort();
            final String image_name_and_version = (clusterResourcesDef.getVersion().length() > 0)
                    ? clusterResourcesDef.getImage() + ":" + clusterResourcesDef.getVersion() : clusterResourcesDef.getImage();

            EnvVar envVar_PREDICTIVE_UNIT_PARAMETERS = new EnvVarBuilder().withName("PREDICTIVE_UNIT_PARAMETERS").withValue(predictiveUnitParameters).build();

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
                    .withEnv(envVar_PREDICTIVE_UNIT_PARAMETERS)
                    .addNewPort().withContainerPort(container_port).endPort()
                    .withNewResources()
                        .addToRequests(resource_requests)
                    .endResources()
                    .build();
            
            containers.add(c);
            //@formatter:on
            
            predictiveUnitIndex++;
        }

        final int replica_number = 1; // clusterResourcesDef.getReplicas();
        final String imagePullSecret = ""; // clusterResourcesDef.getImagePullSecret();
        List<LocalObjectReference> imagePullSecrets = new ArrayList<>();
        if (imagePullSecret.length() > 0) {
            LocalObjectReference imagePullSecretObject = new LocalObjectReference(imagePullSecret);
            imagePullSecrets.add(imagePullSecretObject);
        }

        final String kubernetesDeploymentId = "sd-" + seldonDeploymentId;

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
        return deployment;

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

}
