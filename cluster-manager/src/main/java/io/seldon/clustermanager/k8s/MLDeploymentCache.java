package io.seldon.clustermanager.k8s;

import io.seldon.protos.DeploymentProtos.MLDeployment;

public interface MLDeploymentCache {

	public MLDeployment get(String name);
	public void put(MLDeployment dep);
	public void remove(String name);
	
}
