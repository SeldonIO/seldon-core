package io.seldon.apife.zk;

public interface DeploymentsListener {
	 /**
     * Notification of a change in deployments. DO NOT BLOCK ON THIS METHOD! Long
     * running operations will hold up startup.
     * @param client
     * @param configKey
     * @param configValue
     */
	void deploymentAdded(String deployment, String configValue);
	void deploymentUpdated(String deployment, String configValue);
	void deploymentRemoved(String deployment);
}
