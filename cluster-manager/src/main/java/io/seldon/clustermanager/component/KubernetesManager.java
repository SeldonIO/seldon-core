package io.seldon.clustermanager.component;

import java.util.List;

import io.seldon.protos.DeploymentProtos.DeploymentSpec;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public interface KubernetesManager extends AppComponent {

    public List<String> getNamespaceList();

    public DeploymentSpec createOrReplaceSeldonDeployment(SeldonDeployment mldeployment);

}
