package io.seldon.clustermanager.component;

import io.seldon.protos.DeploymentProtos.CMResultDef;

public interface ClusterManager extends AppComponent {

    public void setZookeeperManager(ZookeeperManager zookeeperManager);
    public void setKubernetesManager(KubernetesManager kubernetesManager);
    
    public CMResultDef getNamespaces();

}
