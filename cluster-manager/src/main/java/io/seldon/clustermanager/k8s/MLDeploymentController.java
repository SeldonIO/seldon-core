package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.MLDeployment;

public interface MLDeploymentController {
	public void createOrReplaceMLDeployment(MLDeployment mlDep);
}
