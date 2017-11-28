package io.seldon.clustermanager.k8s;

import io.kubernetes.client.models.ExtensionsV1beta1DeploymentList;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public interface KubeCRDHandler {

	public void updateSeldonDeployment(SeldonDeployment mlDep);
	public SeldonDeployment getSeldonDeployment(String name);
	public ExtensionsV1beta1DeploymentList getOwnedDeployments(String seldonDeploymentName);
}
