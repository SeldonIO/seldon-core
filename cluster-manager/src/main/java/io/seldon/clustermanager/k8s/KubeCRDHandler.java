package io.seldon.clustermanager.k8s;

import io.fabric8.kubernetes.api.model.OwnerReference;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public interface KubeCRDHandler {

	public void updateMLDeployment(DeploymentDef def,int resourceVersion,OwnerReference oref);
}
