package io.seldon.engine.zk;

import java.util.Map;

public interface DeploymentsHandler {
	Map<String, String> requestCacheDump(String client);

    void addListener(DeploymentsListener listener);

    void addNewDeploymentListener(NewDeploymentListener listener, boolean notifyExistingClients);
}
