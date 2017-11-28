package io.seldon.clustermanager.k8s;

import io.seldon.clustermanager.k8s.SeldonDeploymentOperatorImpl.DeploymentResources;
import io.seldon.protos.DeploymentProtos.SeldonDeployment;

public interface SeldonDeploymentOperator {
	
	public SeldonDeployment defaulting(SeldonDeployment mlDep);
	public void validate(SeldonDeployment mlDep) throws SeldonDeploymentException;
	public DeploymentResources createResources(SeldonDeployment mlDep) throws SeldonDeploymentException;
	public String getKubernetesDeploymentName(String depName,String predictorName);

}
