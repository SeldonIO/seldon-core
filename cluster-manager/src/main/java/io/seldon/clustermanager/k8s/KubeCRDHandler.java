package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public interface KubeCRDHandler {

	public void updateSeldonDeployment(SeldonDeployment mlDep);
	public SeldonDeployment getSeldonDeployment(String name);
}
