package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.DeploymentDef;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public interface KubeCRDHandler {

	public void updateMLDeployment(MLDeployment mlDep);
	public DeploymentDef getMlDeployment(MLDeployment mlDep);
}
