package io.seldon.clustermanager.k8s;

import io.seldon.clustermanager.k8s.MLDeploymentOperatorImpl.DeploymentResources;
import io.seldon.protos.DeploymentProtos.MLDeployment;

public interface MLDeploymentOperator {
	
	public MLDeployment defaulting(MLDeployment mlDep);
	public void validate(MLDeployment mlDep) throws MLDeploymentException;
	public DeploymentResources createResources(MLDeployment mlDep) throws MLDeploymentException;

}
