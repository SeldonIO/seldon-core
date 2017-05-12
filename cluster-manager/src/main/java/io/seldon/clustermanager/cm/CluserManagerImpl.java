package io.seldon.clustermanager.cm;

import java.util.List;
import java.util.Optional;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.component.ZookeeperManager;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.CMStatusDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.DeploymentResultDef;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef;
import io.seldon.protos.DeploymentProtos.StringListDef;

public class CluserManagerImpl implements ClusterManager {

    private final static Logger logger = LoggerFactory.getLogger(CluserManagerImpl.class);

    private ZookeeperManager zookeeperManager;
    private KubernetesManager kubernetesManager;

    public void init() throws Exception {
        logger.info("init");
    }

    public void cleanup() throws Exception {
        logger.info("cleanup");
    }

    @Autowired
    public void setZookeeperManager(ZookeeperManager zookeeperManager) {
        logger.info("injecting ZookeeperManager");
        this.zookeeperManager = zookeeperManager;
    }

    @Autowired
    public void setKubernetesManager(KubernetesManager kubernetesManager) {
        logger.info("injecting KubernetesManager");
        this.kubernetesManager = kubernetesManager;
    }

    @Override
    public CMResultDef getNamespaces() {
        CMResultDef cmResultDef = null;
        try {
            List<String> namespace_list = kubernetesManager.getNamespaceList();

            //@formatter:off
            StringListDef.Builder stringListDefBuilder = StringListDef.newBuilder();
            for (String item: namespace_list) {
                stringListDefBuilder.addItems(item);
            }
            StringListDef stringListDef = stringListDefBuilder.build();
                    
            cmResultDef = CMResultDef.newBuilder()
                    .setCmstatus(CMStatusDef.newBuilder()
                            .setCode(200)
                            .setStatus(CMStatusDef.Status.SUCCESS))
                    .setStringList(stringListDef)
                    .build();
            //@formatter:on

        } catch (Throwable e) {
            logger.error("Error getting namespaces", e);
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            //@formatter:off
            cmResultDef = CMResultDef.newBuilder()
                    .setCmstatus(CMStatusDef.newBuilder()
                            .setCode(500)
                            .setStatus(CMStatusDef.Status.FAILURE)
                            .setInfo(info))
                    .clearStringList()
                    .build();
            //@formatter:on
        }

        return cmResultDef;
    }

    @Override
    public CMResultDef createSeldonDeployment(DeploymentDef deploymentDef) {
        CMResultDef cmResultDef = null;
        try {
            DeploymentDef resultingDeploymentDef = kubernetesManager.createOrReplaceSeldonDeployment(deploymentDef);
            zookeeperManager.persistSeldonDeployment(resultingDeploymentDef);
            //@formatter:off
            DeploymentResultDef deploymentResultDef = DeploymentResultDef.newBuilder()
                    .setDeployment(resultingDeploymentDef)
                    .build();
            //@formatter:on
            cmResultDef = buildSUCCESS(deploymentResultDef);
        } catch (Throwable e) {
            logger.error("Error creating seldon deployment", e);
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE_500(info);
        }
        return cmResultDef;
    }

    @Override
    public CMResultDef getSeldonDeployment(DeploymentDef deploymentDef) {
        CMResultDef cmResultDef = null;
        try {
            Optional<DeploymentDef> deploymentDefInStorage = zookeeperManager.getSeldonDeployment(deploymentDef);
            if (deploymentDefInStorage.isPresent()) {
                DeploymentDef resultingDeploymentDef = kubernetesManager.getSeldonDeployment(deploymentDefInStorage.get());
                //@formatter:off
                DeploymentResultDef deploymentResultDef = DeploymentResultDef.newBuilder()
                        .setDeployment(resultingDeploymentDef)
                        .build();
                //@formatter:on
                cmResultDef = buildSUCCESS(deploymentResultDef);
            } else {
                final String seldonDeploymentId = deploymentDef.getId();
                final String info = String.format("seldonDeploymentId[%s] not found in zookeeper", seldonDeploymentId);
                cmResultDef = buildFAILURE_404(info);
            }
        } catch (Throwable e) {
            logger.error("Error getting seldon deployment", e);
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE_500(info);
        }
        return cmResultDef;
    }

    @Override
    public CMResultDef updateSeldonDeployment(DeploymentDef deploymentDef) {
        return this.createSeldonDeployment(deploymentDef);
    }

    @Override
    public CMResultDef deleteSeldonDeployment(DeploymentDef deploymentDef) {
        CMResultDef cmResultDef = null;
        try {
            kubernetesManager.deleteSeldonDeployment(deploymentDef);
            zookeeperManager.deleteSeldonDeployment(deploymentDef);
            cmResultDef = buildSUCCESS();
        } catch (Throwable e) {
            logger.error("Error deleting seldon deployment", e);
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE_500(info);
        }
        return cmResultDef;
    }

    @Override
    public CMResultDef createOrReplaceDockerRegistrySecret(DockerRegistrySecretDef dockerRegistrySecretDef) {
        CMResultDef cmResultDef = null;
        try {
            kubernetesManager.createOrReplaceDockerRegistrySecret(dockerRegistrySecretDef);
            cmResultDef = buildSUCCESS();
        } catch (Throwable e) {
            logger.error("Error creating/updating docker registry secret", e);
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE_500(info);
        }
        return cmResultDef;
    }

    @Override
    public CMResultDef deleteDockerRegistrySecret(String name) {
        CMResultDef cmResultDef = null;
        try {
            kubernetesManager.deleteDockerRegistrySecret(name);
            cmResultDef = buildSUCCESS();
        } catch (Throwable e) {
            logger.error("Error deleting docker registry secret", e);
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE_500(info);
        }
        return cmResultDef;
    }

    private static CMResultDef buildFAILURE(String info, int code) {
        //@formatter:off
        CMResultDef cmResultDef = CMResultDef.newBuilder()
                .setCmstatus(CMStatusDef.newBuilder()
                        .setCode(code)
                        .setStatus(CMStatusDef.Status.FAILURE)
                        .setInfo(info))
                .clearOneofData()
                .build();
        //@formatter:on
        return cmResultDef;
    }

    private static CMResultDef buildFAILURE_404(String info) {
        return buildFAILURE(info, 404);
    }

    private static CMResultDef buildFAILURE_500(String info) {
        return buildFAILURE(info, 500);
    }

    private static CMResultDef buildSUCCESS() {
        //@formatter:off
        CMResultDef cmResultDef = CMResultDef.newBuilder()
                .setCmstatus(CMStatusDef.newBuilder()
                        .setCode(200)
                        .setStatus(CMStatusDef.Status.SUCCESS))
                .clearOneofData()
                .build();
        //@formatter:on
        return cmResultDef;
    }

    private static CMResultDef buildSUCCESS(DeploymentResultDef deploymentResultDef) {
        //@formatter:off
        CMResultDef cmResultDef = CMResultDef.newBuilder()
                .setCmstatus(CMStatusDef.newBuilder()
                        .setCode(200)
                        .setStatus(CMStatusDef.Status.SUCCESS))
                .setDeploymentResult(deploymentResultDef)
                .build();
        //@formatter:on
        return cmResultDef;
    }

}
