package io.seldon.clustermanager.k8s;

import java.util.Optional;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.fabric8.kubernetes.api.model.extensions.DeploymentBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;

class KubernetesDeploymentOps {

    private final static Logger logger = LoggerFactory.getLogger(KubernetesDeploymentOps.class);

    private final String namespace_name;
    private final String seldonDeploymentId;
    private final KubernetesClient kubernetesClient;

    public KubernetesDeploymentOps(String seldonDeploymentId, KubernetesClient kubernetesClient, String namespace_name) {
        this.seldonDeploymentId = seldonDeploymentId;
        this.kubernetesClient = kubernetesClient;
        this.namespace_name = namespace_name;
    }

    public Optional<Deployment> create(String kubernetesDeploymentId, ClusterResourcesDef clusterResourcesDef) {

        Optional<Deployment> retVal = Optional.empty();
        if (hasDeployableImage(clusterResourcesDef)) {
            final int replica_number = clusterResourcesDef.getReplicas();
            final int container_port = 80; // TODO change this!
            final String image_name_and_version = clusterResourcesDef.getImage() + ":" + clusterResourcesDef.getVersion();

            //@formatter:off
            Deployment deployment = new DeploymentBuilder()
                .withNewMetadata().withName(kubernetesDeploymentId).addToLabels("seldon-deployment-id", seldonDeploymentId).endMetadata()
                .withNewSpec().withReplicas(replica_number)
                    .withNewTemplate()
                        .withNewMetadata().addToLabels("seldon-app", kubernetesDeploymentId).endMetadata()
                        .withNewSpec().addNewContainer().withName("seldon-container").withImage(image_name_and_version)
                            .addNewPort().withContainerPort(container_port).endPort().endContainer()
                        .endSpec()
                    .endTemplate()
                .endSpec().build();
            //@formatter:on

            deployment = kubernetesClient.extensions().deployments().inNamespace(namespace_name).create(deployment);
            String deploymentName = (deployment != null) ? deployment.getMetadata().getName() : null;
            logger.debug(String.format("Created kubernetes delployment [%s]", deploymentName));
            retVal = Optional.of(deployment);
        } else {
            logger.debug(String.format("Ignoring kubernetes delployment [%s], not deployable", kubernetesDeploymentId));
        }

        return retVal;
    }

    public Deployment update(String kubernetesDeploymentId, ClusterResourcesDef clusterResourcesDef) {
        final int replica_number = clusterResourcesDef.getReplicas();
        final int container_port = 80; // TODO change this!
        final String image_name_and_version = clusterResourcesDef.getImage() + ":" + clusterResourcesDef.getVersion();

        //@formatter:off
        Deployment deployment = new DeploymentBuilder()
                .withNewMetadata().withName(kubernetesDeploymentId).addToLabels("seldon-deployment-id", seldonDeploymentId).endMetadata()
                .withNewSpec().withReplicas(replica_number)
                    .withNewTemplate()
                        .withNewMetadata().addToLabels("seldon-app", kubernetesDeploymentId).endMetadata()
                        .withNewSpec().addNewContainer().withName("seldon-container").withImage(image_name_and_version)
                            .addNewPort().withContainerPort(container_port).endPort().endContainer()
                        .endSpec()
                    .endTemplate()
                .endSpec().build();
            //@formatter:on

        deployment = kubernetesClient.extensions().deployments().inNamespace(namespace_name).createOrReplace(deployment);
        String deploymentName = (deployment != null) ? deployment.getMetadata().getName() : null;
        logger.debug(String.format("Updated kubernetes delployment [%s]", deploymentName));

        return deployment;
    }

    public void delete(Deployment deployment) {
        final String deploymentName = deployment.getMetadata().getName();

        io.fabric8.kubernetes.api.model.extensions.ReplicaSetList rslist = kubernetesClient.extensions().replicaSets().inNamespace(namespace_name)
                .withLabel("seldon-app", deploymentName).list();
        kubernetesClient.resource(deployment).delete();
        logger.debug(String.format("Deleted kubernetes delployment [%s]", deploymentName));
        for (io.fabric8.kubernetes.api.model.extensions.ReplicaSet rs : rslist.getItems()) {
            kubernetesClient.resource(rs).delete();
            String rsmsg = (rs != null) ? rs.getMetadata().getName() : null;
            logger.debug(String.format("Deleted kubernetes replicaSet [%s]", rsmsg));
        }
    }

    public static boolean hasDeployableImage(ClusterResourcesDef clusterResourcesDef) {
        return (clusterResourcesDef.getImage().length() > 0);
    }

}
