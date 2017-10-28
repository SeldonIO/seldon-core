package io.seldon.clustermanager.k8s;

import java.util.ArrayList;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceList;
import io.fabric8.kubernetes.client.Config;
import io.fabric8.kubernetes.client.ConfigBuilder;
import io.fabric8.kubernetes.client.DefaultKubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.fabric8.kubernetes.client.KubernetesClientException;
import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.DeploymentUtils.BuildDeploymentResult;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class KubernetesManagerImpl implements KubernetesManager {

    private final static Logger logger = LoggerFactory.getLogger(KubernetesManagerImpl.class);
    private final static String SELDON_CLUSTER_MANAGER_POD_NAMESPACE_KEY = "SELDON_CLUSTER_MANAGER_POD_NAMESPACE";

    private ClusterManagerProperites clusterManagerProperites;
    private KubernetesClient kubernetesClient = null;
    private KubeCRDHandler kubeCRDHandler;
    
    private String seldonClusterNamespaceName = "UNKOWN_NAMESPACE";

    public KubernetesManagerImpl() {
    }

    @Override
    public void init() throws Exception {
        logger.info("init");

        Config config = new ConfigBuilder().build(); // use defaults for config
        logger.info(String.format("Connecting to kubernetes master[%s]", config.getMasterUrl()));
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

    @Autowired
    public void setClusterManagerProperites(ClusterManagerProperites clusterManagerProperites) {
        logger.info(String.format("injecting %s", clusterManagerProperites.toString()));
        this.clusterManagerProperites = clusterManagerProperites;
    }

    @Autowired
    public void setKubeCRDHandler(KubeCRDHandler kubeCRDHandler) {
    	logger.info("Injecting KubeCRDHandler");
    	this.kubeCRDHandler = kubeCRDHandler;
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
    public DeploymentDef createOrReplaceSeldonDeployment(DeploymentDef deploymentDef,CustomResourceDetails crd) {
        DeploymentDef.Builder resultingDeploymentDefBuilder = DeploymentDef.newBuilder(deploymentDef);
        final String seldonDeploymentId = deploymentDef.getId();
        logger.debug(String.format("Creating Seldon Deployment id[%s]", seldonDeploymentId));
        final String namespace_name = getNamespaceName();

        List<BuildDeploymentResult> deploymentResult = DeploymentUtils.buildDeployments(deploymentDef, clusterManagerProperites,crd.getOref());
        
        
        deploymentResult.stream().forEach((buildDeploymentResult) -> {

            { // update the resultingDeploymentDef with the predictor having the predictive unit endpoints
                if (buildDeploymentResult.isCanary) {
                    resultingDeploymentDefBuilder.setPredictorCanary(buildDeploymentResult.resultingPredictorDef);
                } else {
                    resultingDeploymentDefBuilder.setPredictor(buildDeploymentResult.resultingPredictorDef);
                }
            }
        });

        DeploymentDef resultingDeploymentDef = resultingDeploymentDefBuilder.build();
        
        if (!resultingDeploymentDef.equals(deploymentDef))
        {
        	logger.info("Updating ML Deployment resource");
        	kubeCRDHandler.updateMLDeployment(resultingDeploymentDef, crd);
        }
        else
        {

        	deploymentResult.stream().forEach((buildDeploymentResult) -> {
        		DeploymentUtils.createDeployment(kubernetesClient, namespace_name, buildDeploymentResult);
        	});
        	
        	// remove a canary if necessary
        	if (!deploymentDef.hasField(deploymentDef.getDescriptorForType().findFieldByNumber(DeploymentDef.PREDICTOR_CANARY_FIELD_NUMBER))) {
        		DeploymentUtils.deleteDeployemntResources(kubernetesClient, namespace_name, seldonDeploymentId, true);
        	}
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


    private String getNamespaceName() {
        return seldonClusterNamespaceName;
    }
}
