package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.DeploymentDef;

public interface KubeCRDHandler {

	public void updateMLDeployment(DeploymentDef def,CustomResourceDetails crd);
	public DeploymentDef getMlDeployment(CustomResourceDetails crd);
}
