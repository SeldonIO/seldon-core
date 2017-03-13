package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.Base64;
import java.util.Collection;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Collectors;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceList;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.fabric8.kubernetes.api.model.extensions.DeploymentList;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientException;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.protos.DeploymentProtos.ClusterResourcesDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef;
import io.seldon.protos.DeploymentProtos.EndpointDef;
import io.seldon.protos.DeploymentProtos.PredictiveUnitDef;
import io.seldon.protos.DeploymentProtos.PredictorDef;
import io.seldon.protos.DeploymentProtos.StringSecretDef;

public class KubernetesManagerImpl implements KubernetesManager {

    private final static Logger logger = LoggerFactory.getLogger(KubernetesManagerImpl.class);

    private KubernetesClient kubernetesClient = null;

    @Override
    public void init() throws Exception {
        logger.info("init");

        String master = "http://localhost:8001/";
        Config config = new ConfigBuilder().withMasterUrl(master).build();

        try {
            kubernetesClient = new DefaultKubernetesClient(config);
            getNamespaceList(); // simple check to see if client works
            logger.info("Sucessfully passed namespace check");
        } catch (KubernetesClientException e) {
            throw new Exception(e);
        }
    }

    @Override
    public void cleanup() throws Exception {
        logger.info("cleanup");
        if (kubernetesClient != null) {
            kubernetesClient.close();
        }
    }

    public List<String> getNamespaceList() {
        List<String> namespace_list = new ArrayList<>();
        NamespaceList namespaceList = kubernetesClient.namespaces().list();
        for (Namespace ns : namespaceList.getItems()) {
            namespace_list.add(ns.getMetadata().getName());
        }

        return namespace_list;
    }

    @Override
    public DeploymentDef createSeldonDeployment(DeploymentDef deploymentDef) {
        DeploymentDef.Builder resultingDeploymentDefBuilder = DeploymentDef.newBuilder(deploymentDef);
        final String seldonDeploymentId = Long.toString(deploymentDef.getId());
        logger.debug(String.format("Creating Seldon Deployment id[%s]", seldonDeploymentId));
        final String namespace_name = "default"; // TODO change this!

        PredictorDef predictorDef = deploymentDef.getPredictor();

        List<PredictiveUnitDef> predictiveUnits = predictorDef.getPredictiveUnitsList();
        int predictiveUnitIndex = 0;
        for (PredictiveUnitDef predictiveUnitDef : predictiveUnits) {
            final String predictiveUnitId = Long.toString(predictiveUnitDef.getId());
            final String predictive_unit_name = predictiveUnitDef.getName();
            final String kubernetesDeploymentId = getKubernetesDeploymentId(seldonDeploymentId, predictiveUnitId);
            logger.debug(String.format("Deploying predictiveUnit[%s] for seldonDeployment id[%s]", predictive_unit_name, seldonDeploymentId));

            ClusterResourcesDef clusterResourcesDef = predictiveUnitDef.getClusterResources();
            EndpointDef endpointDef = predictiveUnitDef.getEndpoint();
            Optional<Deployment> deployment = new KubernetesDeploymentOps(seldonDeploymentId, kubernetesClient, namespace_name).create(kubernetesDeploymentId,
                    clusterResourcesDef, endpointDef);
            if (deployment.isPresent()) {
                Service service = new KubernetesServiceOps(kubernetesClient, namespace_name, deployment.get()).create(endpointDef);
                /// String serviceClusterIP = service.getSpec().getClusterIP();
                String serviceName = service.getMetadata().getName();
                resultingDeploymentDefBuilder.getPredictorBuilder().getPredictiveUnitsBuilder(predictiveUnitIndex).getEndpointBuilder()
                        .setServiceHost(serviceName);
            }

            predictiveUnitIndex++;
        }

        return resultingDeploymentDefBuilder.build();
    }

    @Override
    public DeploymentDef updateSeldonDeployment(DeploymentDef deploymentDef) {
        DeploymentDef.Builder resultingDeploymentDefBuilder = DeploymentDef.newBuilder(deploymentDef);
        final String seldonDeploymentId = Long.toString(deploymentDef.getId());
        logger.debug(String.format("Updating Seldon Deployment id[%s]", seldonDeploymentId));
        final String namespace_name = "default"; // TODO change this!

        Set<String> requiredDeployments = new HashSet<>();
        { // check required deployment list
            PredictorDef predictorDef = deploymentDef.getPredictor();
            List<PredictiveUnitDef> predictiveUnits = predictorDef.getPredictiveUnitsList();
            for (PredictiveUnitDef predictiveUnitDef : predictiveUnits) {
                final String predictiveUnitId = Long.toString(predictiveUnitDef.getId());
                final String kubernetesDeploymentId = getKubernetesDeploymentId(seldonDeploymentId, predictiveUnitId);
                requiredDeployments.add(kubernetesDeploymentId);
            }
        }
        Set<String> existingDeployments = new HashSet<>();
        { // find existing deployments
            DeploymentList deployments = kubernetesClient.extensions().deployments().inNamespace(namespace_name)
                    .withLabel("seldon-deployment-id", seldonDeploymentId).list();
            for (Deployment deployment : deployments.getItems()) {
                String kubernetesDeploymentId = deployment.getMetadata().getName();
                existingDeployments.add(kubernetesDeploymentId);
            }
        }

        { // Delete deployments not required anymore
            DeploymentList deployments = kubernetesClient.extensions().deployments().inNamespace(namespace_name)
                    .withLabel("seldon-deployment-id", seldonDeploymentId).list();
            for (Deployment deployment : deployments.getItems()) {
                String kubernetesDeploymentId = deployment.getMetadata().getName();
                if (!requiredDeployments.contains(kubernetesDeploymentId)) {
                    new KubernetesDeploymentOps(seldonDeploymentId, kubernetesClient, namespace_name).delete(deployment);
                    new KubernetesServiceOps(kubernetesClient, namespace_name, deployment).delete();
                }
            }

        }

        { // Update or create the required deployments
            PredictorDef predictorDef = deploymentDef.getPredictor();
            List<PredictiveUnitDef> predictiveUnits = predictorDef.getPredictiveUnitsList();
            int predictiveUnitIndex = 0;
            for (PredictiveUnitDef predictiveUnitDef : predictiveUnits) {
                final String predictiveUnitId = Long.toString(predictiveUnitDef.getId());
                final String kubernetesDeploymentId = getKubernetesDeploymentId(seldonDeploymentId, predictiveUnitId);
                final String predictive_unit_name = predictiveUnitDef.getName();
                logger.debug(String.format("Deploying predictiveUnit[%s] for seldonDeployment id[%s]", predictive_unit_name, seldonDeploymentId));
                ClusterResourcesDef clusterResourcesDef = predictiveUnitDef.getClusterResources();
                EndpointDef endpointDef = predictiveUnitDef.getEndpoint();
                if (existingDeployments.contains(kubernetesDeploymentId)) {
                    Deployment deployment = new KubernetesDeploymentOps(seldonDeploymentId, kubernetesClient, namespace_name).update(kubernetesDeploymentId,
                            clusterResourcesDef, endpointDef);
                    Service service = new KubernetesServiceOps(kubernetesClient, namespace_name, deployment).update(endpointDef);
                    String serviceName = service.getMetadata().getName();
                    resultingDeploymentDefBuilder.getPredictorBuilder().getPredictiveUnitsBuilder(predictiveUnitIndex).getEndpointBuilder()
                            .setServiceHost(serviceName);
                } else {
                    Optional<Deployment> deployment = new KubernetesDeploymentOps(seldonDeploymentId, kubernetesClient, namespace_name)
                            .create(kubernetesDeploymentId, clusterResourcesDef, endpointDef);
                    if (deployment.isPresent()) {
                        Service service = new KubernetesServiceOps(kubernetesClient, namespace_name, deployment.get()).create(endpointDef);
                        String serviceName = service.getMetadata().getName();
                        resultingDeploymentDefBuilder.getPredictorBuilder().getPredictiveUnitsBuilder(predictiveUnitIndex).getEndpointBuilder()
                                .setServiceHost(serviceName);
                    }
                }

                predictiveUnitIndex++;
            }

        }

        return resultingDeploymentDefBuilder.build();
    }

    @Override
    public void deleteSeldonDeployment(DeploymentDef deploymentDef) {
        final String seldonDeploymentId = Long.toString(deploymentDef.getId());
        logger.debug(String.format("Deleting Seldon Deployment[%s]", seldonDeploymentId));
        final String namespace_name = "default"; // TODO change this!

        DeploymentList deployments = kubernetesClient.extensions().deployments().inNamespace(namespace_name)
                .withLabel("seldon-deployment-id", seldonDeploymentId).list();
        for (Deployment deployment : deployments.getItems()) {
            new KubernetesDeploymentOps(seldonDeploymentId, kubernetesClient, namespace_name).delete(deployment);
            new KubernetesServiceOps(kubernetesClient, namespace_name, deployment).delete();
        }

    }

    /**
     * Convert a SeldonDeploymentId into a KubernetesDeploymentId
     */
    private static String getKubernetesDeploymentId(String seldonDeploymentId, String predictiveUnitId) {
        return "sd-" + seldonDeploymentId + "-" + predictiveUnitId;
    }

    @Override
    public void createOrReplaceStringSecret(StringSecretDef stringSecretDef) {
        final String namespace_name = "default"; // TODO change this!
        Secret secret = new KubernetesSecretOps(kubernetesClient, namespace_name).createOrReplaceSecret(stringSecretDef);
    }

    @Override
    public void deleteStringSecret(String name) {
        final String namespace_name = "default"; // TODO change this!
        new KubernetesSecretOps(kubernetesClient, namespace_name).deleteSecret(name);
    }

    @Override
    public void createOrReplaceDockerRegistrySecret(DockerRegistrySecretDef dockerRegistrySecretDef) {
        final String namespace_name = "default"; // TODO change this!
        new KubernetesSecretOps(kubernetesClient, namespace_name).createOrReplaceSecret(dockerRegistrySecretDef);
    }

    @Override
    public void deleteDockerRegistrySecret(String name) {
        final String namespace_name = "default"; // TODO change this!
        new KubernetesSecretOps(kubernetesClient, namespace_name).deleteSecret(name);
    }

}
