package io.seldon.clustermanager.component;

import java.util.List;

import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.StringSecretDef;

public interface KubernetesManager extends AppComponent {

    public List<String> getNamespaceList();

    public DeploymentDef createSeldonDeployment(DeploymentDef deploymentDef);
    public DeploymentDef updateSeldonDeployment(DeploymentDef deploymentDef);
    public void deleteSeldonDeployment(DeploymentDef deploymentDef);
    public void createStringSecret(StringSecretDef stringSecretDef);
    public void deleteStringSecret(String name);
}
