package io.seldon.clustermanager.component;

import io.seldon.protos.DeploymentProtos.DeploymentDef;

public interface ZookeeperManager extends AppComponent {

    public void persistSeldonDeployment(DeploymentDef deploymentDef) throws Exception;
    public void deleteSeldonDeployment(DeploymentDef deploymentDef) throws Exception;
    
}
