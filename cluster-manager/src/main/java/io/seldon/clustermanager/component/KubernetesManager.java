package io.seldon.clustermanager.component;

import java.util.List;

import io.seldon.clustermanager.k8s.CustomResourceDetails;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public interface KubernetesManager extends AppComponent {

    public List<String> getNamespaceList();

    public DeploymentDef createOrReplaceSeldonDeployment(DeploymentDef deploymentDef,CustomResourceDetails crd);

    public DeploymentDef getSeldonDeployment(DeploymentDef deploymentDef);

    public void deleteSeldonDeployment(DeploymentDef deploymentDef);

}
