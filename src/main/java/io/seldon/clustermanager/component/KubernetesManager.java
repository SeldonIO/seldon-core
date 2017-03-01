package io.seldon.clustermanager.component;

import java.util.List;

import io.seldon.protos.DeploymentProtos.CMResultDef;

public interface KubernetesManager extends AppComponent {

    public List<String> getNamespaceList();

}
