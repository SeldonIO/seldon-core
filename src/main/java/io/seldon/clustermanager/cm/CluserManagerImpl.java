package io.seldon.clustermanager.cm;

import java.util.List;

import org.springframework.beans.factory.annotation.Autowired;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.component.ZookeeperManager;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.CMStatusDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.StringListDef;

public class CluserManagerImpl implements ClusterManager {

    private ZookeeperManager zookeeperManager;
    private KubernetesManager kubernetesManager;

    public void init() throws Exception {
        System.out.println("ClusterManager: init");
    }

    public void cleanup() throws Exception {
        System.out.println("ClusterManager: cleanup");
    }

    @Autowired
    public void setZookeeperManager(ZookeeperManager zookeeperManager) {
        System.out.println("ClusterManager: set ZookeeperManager injection");
        this.zookeeperManager = zookeeperManager;
    }

    @Autowired
    public void setKubernetesManager(KubernetesManager kubernetesManager) {
        System.out.println("ClusterManager: set KubernetesManager injection");
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
            kubernetesManager.createSeldonDeployment(deploymentDef);
            cmResultDef = buildSUCCESS();
        } catch (Throwable e) {
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE(info);
        }
        return cmResultDef;
    }

    @Override
    public CMResultDef updateSeldonDeployment(DeploymentDef deploymentDef) {
        CMResultDef cmResultDef = null;
        try {
            kubernetesManager.updateSeldonDeployment(deploymentDef);
            cmResultDef = buildSUCCESS();
        } catch (Throwable e) {
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE(info);
        }
        return cmResultDef;
    }

    @Override
    public CMResultDef deleteSeldonDeployment(DeploymentDef deploymentDef) {
        CMResultDef cmResultDef = null;
        try {
            kubernetesManager.deleteSeldonDeployment(deploymentDef);
            cmResultDef = buildSUCCESS();
        } catch (Throwable e) {
            String info = org.apache.commons.lang3.exception.ExceptionUtils.getStackTrace(e);
            cmResultDef = buildFAILURE(info);
        }
        return cmResultDef;
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

    private static CMResultDef buildFAILURE(String info) {
        //@formatter:off
        CMResultDef cmResultDef = CMResultDef.newBuilder()
                .setCmstatus(CMStatusDef.newBuilder()
                        .setCode(500)
                        .setStatus(CMStatusDef.Status.FAILURE)
                        .setInfo(info))
                .clearOneofData()
                .build();
        //@formatter:on
        return cmResultDef;
    }
}
