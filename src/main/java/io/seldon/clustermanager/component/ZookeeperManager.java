package io.seldon.clustermanager.component;

import io.seldon.protos.DeploymentProtos.DeploymentDef;

public interface ZookeeperManager extends AppComponent {

    public void persistDeployment(DeploymentDef deploymentDef) throws Exception;
    public void deleteDeployment(DeploymentDef deploymentDef) throws Exception;
    
}
