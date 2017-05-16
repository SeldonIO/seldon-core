package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceList;
import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientException;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef;
import io.seldon.protos.DeploymentProtos.StringSecretDef;

public class KubernetesManagerImpl implements KubernetesManager {

    private final static Logger logger = LoggerFactory.getLogger(KubernetesManagerImpl.class);
    private final static String SELDON_CLUSTER_MANAGER_POD_NAMESPACE_KEY = "SELDON_CLUSTER_MANAGER_POD_NAMESPACE";

    private KubernetesClient kubernetesClient = null;

    private String seldonClusterNamespaceName = "UNKOWN_NAMESPACE";

    public KubernetesManagerImpl() {
    }

    @Override
    public void init() throws Exception {
        logger.info("init");

        String master = "http://localhost:8001/";
        logger.info(String.format("Connecting to kubernetes master[%s]", master));
        Config config = new ConfigBuilder().withMasterUrl(master).build();

        try {
            kubernetesClient = new DefaultKubernetesClient(config);
            getNamespaceList(); // simple check to see if client works
            logger.info("Sucessfully passed getNamespaceList() check");
        } catch (KubernetesClientException e) {
            throw new Exception(e);
        }

        { // set the namespace to use
            seldonClusterNamespaceName = System.getenv().get(SELDON_CLUSTER_MANAGER_POD_NAMESPACE_KEY);
            if (seldonClusterNamespaceName == null) {
                logger.error(String.format("FAILED to find env var [%s]", SELDON_CLUSTER_MANAGER_POD_NAMESPACE_KEY));
                seldonClusterNamespaceName = "default";
            }
            logger.info(String.format("Setting cluster manager namespace as [%s]", seldonClusterNamespaceName));
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
    public DeploymentDef createOrReplaceSeldonDeployment(DeploymentDef deploymentDef) {
        DeploymentDef.Builder resultingDeploymentDefBuilder = DeploymentDef.newBuilder(deploymentDef);
        final String seldonDeploymentId = deploymentDef.getId();
        logger.debug(String.format("Creating Seldon Deployment id[%s]", seldonDeploymentId));
        final String namespace_name = getNamespaceName();

        DeploymentUtils.buildDeployments(deploymentDef).stream().forEach((buildDeploymentResult) -> {
            DeploymentUtils.createDeployment(kubernetesClient, namespace_name, buildDeploymentResult);
            { // update the resultingDeploymentDef with the predictor having the predictive unit endpoints
                if (buildDeploymentResult.isCanary) {
                    resultingDeploymentDefBuilder.setPredictorCanary(buildDeploymentResult.resultingPredictorDef);
                } else {
                    resultingDeploymentDefBuilder.setPredictor(buildDeploymentResult.resultingPredictorDef);
                }
            }
        });

        // remove a canary if necessary
        if (!deploymentDef.hasField(deploymentDef.getDescriptorForType().findFieldByNumber(DeploymentDef.PREDICTOR_CANARY_FIELD_NUMBER))) {
            DeploymentUtils.deleteDeployemntResources(kubernetesClient, namespace_name, seldonDeploymentId, true);
        }

        return resultingDeploymentDefBuilder.build();

    }

    @Override
    public DeploymentDef getSeldonDeployment(DeploymentDef deploymentDef) {
        final String seldonDeploymentId = deploymentDef.getId();
        logger.debug(String.format("Getting Seldon Deployment id[%s]", seldonDeploymentId));
        final String namespace_name = getNamespaceName();
        DeploymentDef resultingDeploymentDef = DeploymentUtils.getDeployments(kubernetesClient, namespace_name, deploymentDef);
        return resultingDeploymentDef;
    }

    @Override
    public void deleteSeldonDeployment(DeploymentDef deploymentDef) {
        final String seldonDeploymentId = deploymentDef.getId();
        logger.debug(String.format("Deleting Seldon Deployment id[%s]", seldonDeploymentId));
        String namespace_name = getNamespaceName();
        DeploymentUtils.deleteDeployment(kubernetesClient, namespace_name, deploymentDef);
    }

    @Override
    public void createOrReplaceStringSecret(StringSecretDef stringSecretDef) {
        final String namespace_name = getNamespaceName();
        Secret secret = SecretUtils.createOrReplaceSecret(kubernetesClient, namespace_name, stringSecretDef);
    }

    @Override
    public void deleteStringSecret(String name) {
        final String namespace_name = getNamespaceName();
        SecretUtils.deleteSecret(kubernetesClient, namespace_name, name);
    }

    @Override
    public void createOrReplaceDockerRegistrySecret(DockerRegistrySecretDef dockerRegistrySecretDef) {
        final String namespace_name = getNamespaceName();
        SecretUtils.createOrReplaceSecret(kubernetesClient, namespace_name, dockerRegistrySecretDef);
    }

    @Override
    public void deleteDockerRegistrySecret(String name) {
        final String namespace_name = getNamespaceName();
        SecretUtils.deleteSecret(kubernetesClient, namespace_name, name);
    }

    private String getNamespaceName() {
        return seldonClusterNamespaceName;
    }
}
