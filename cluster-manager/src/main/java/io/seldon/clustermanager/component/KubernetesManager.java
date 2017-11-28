package io.seldon.clustermanager.component;

import java.util.List;

import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public interface KubernetesManager extends AppComponent {

    public List<String> getNamespaceList();

    public DeploymentDef createOrReplaceSeldonDeployment(MLDeployment mldeployment);

}
