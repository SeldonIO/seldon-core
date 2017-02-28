package io.seldon.clustermanager.component;

import io.seldon.protos.DeploymentProtos.CMResultDef;

public interface KubernetesManager extends AppComponent {

    public CMResultDef getNamespaces();
    
}
