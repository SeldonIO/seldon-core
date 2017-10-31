package io.seldon.apife.deployments;

import io.seldon.protos.DeploymentProtos.MLDeployment;

public interface DeploymentsListener {
	 /**
     * Notification of a change in deployments. DO NOT BLOCK ON THIS METHOD! Long
     * running operations will hold up startup.
     * @param client
     * @param configKey
     * @param configValue
     */
	void deploymentAdded(MLDeployment resource);
	void deploymentUpdated(MLDeployment resource);
	void deploymentRemoved(MLDeployment resource);
}
