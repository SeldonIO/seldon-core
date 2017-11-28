package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public interface SeldonDeploymentCache {

	public SeldonDeployment get(String name);
	public void put(SeldonDeployment dep);
	public void remove(String name);
	
}
