package io.seldon.apife.deployments;

public interface DeploymentsListener {
	 /**
     * Notification of a change in deployments. DO NOT BLOCK ON THIS METHOD! Long
     * running operations will hold up startup.
     * @param client
     * @param configKey
     * @param configValue
     */
	void deploymentAdded(String resource);
	void deploymentUpdated(String resource);
	void deploymentRemoved(String resource);
}
