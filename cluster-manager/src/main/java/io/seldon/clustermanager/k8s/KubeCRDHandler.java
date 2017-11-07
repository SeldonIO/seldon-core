package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.MLDeployment;

public interface KubeCRDHandler {

	public void updateMLDeployment(MLDeployment mlDep);
	public MLDeployment getMlDeployment(String name);
}
