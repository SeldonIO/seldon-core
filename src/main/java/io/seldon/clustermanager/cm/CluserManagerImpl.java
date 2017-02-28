package io.seldon.clustermanager.cm;

import org.springframework.beans.factory.annotation.Autowired;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.component.ZookeeperManager;

public class CluserManagerImpl implements ClusterManager {

    private ZookeeperManager zookeeperManager;
    private KubernetesManager kubernetesManager;

    public void cleanup() throws Exception {
        System.out.println("ClusterManager cleanup");
    }

    public void init() throws Exception {
        System.out.println("ClusterManager init");
    }

    @Autowired
    public void setZookeeperManager(ZookeeperManager zookeeperManager) {
        System.out.println("ClusterManager set ZookeeperManager injection");
        this.zookeeperManager = zookeeperManager;
    }

    @Autowired
    public void setKubernetesManager(KubernetesManager kubernetesManager) {
        System.out.println("ClusterManager set KubernetesManager injection");
        this.kubernetesManager = kubernetesManager;
    }

}
