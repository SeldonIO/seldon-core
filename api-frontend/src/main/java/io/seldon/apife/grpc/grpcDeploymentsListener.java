package io.seldon.apife.grpc;

import io.seldon.apife.deployments.DeploymentsListener;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public class grpcDeploymentsListener implements DeploymentsListener {

    private final SeldonGrpcServer server;
        
    public grpcDeploymentsListener(SeldonGrpcServer server) {
        super();
        this.server = server;
     }

    @Override
    public void deploymentAdded(SeldonDeployment resource) {
        server.deploymentAdded(resource);
    }

    @Override
    public void deploymentUpdated(SeldonDeployment resource) {
        // Do nothing
    }

    @Override
    public void deploymentRemoved(SeldonDeployment resource) {
       server.deploymentRemoved(resource);
    }
}
