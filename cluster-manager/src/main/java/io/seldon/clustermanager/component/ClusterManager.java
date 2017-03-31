package io.seldon.clustermanager.component;

import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef;

public interface ClusterManager extends AppComponent {

    public void setZookeeperManager(ZookeeperManager zookeeperManager);

    public void setKubernetesManager(KubernetesManager kubernetesManager);

    public CMResultDef getNamespaces();

    public CMResultDef createSeldonDeployment(DeploymentDef deploymentDef);

    public CMResultDef getSeldonDeployment(DeploymentDef deploymentDef);

    public CMResultDef updateSeldonDeployment(DeploymentDef deploymentDef);

    public CMResultDef deleteSeldonDeployment(DeploymentDef deploymentDef);

    public CMResultDef createOrReplaceDockerRegistrySecret(DockerRegistrySecretDef dockerRegistrySecretDef);

    public CMResultDef deleteDockerRegistrySecret(String name);

}
