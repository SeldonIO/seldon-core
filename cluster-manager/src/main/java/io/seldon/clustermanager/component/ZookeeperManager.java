package io.seldon.clustermanager.component;

import java.util.Optional;

import io.seldon.protos.DeploymentProtos.DeploymentDef;

public interface ZookeeperManager extends AppComponent {

    public void persistSeldonDeployment(DeploymentDef deploymentDef) throws Exception;

    public Optional<DeploymentDef> getSeldonDeployment(DeploymentDef deploymentDef) throws Exception;

    public void deleteSeldonDeployment(DeploymentDef deploymentDef) throws Exception;

}
