package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public interface SeldonDeploymentController {
	public void createOrReplaceMLDeployment(SeldonDeployment mlDep);
}
