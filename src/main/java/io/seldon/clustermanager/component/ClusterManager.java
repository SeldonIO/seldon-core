package io.seldon.clustermanager.component;

public interface ClusterManager extends AppComponent {

    public void setZookeeperManager(ZookeeperManager zookeeperManager);
    public void setKubernetesManager(KubernetesManager kubernetesManager);
}
