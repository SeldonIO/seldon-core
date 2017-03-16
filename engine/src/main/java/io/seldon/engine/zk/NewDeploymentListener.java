package io.seldon.engine.zk;

import java.util.Map;

public interface NewDeploymentListener {
	void deploymentAdded(String deployment, Map<String, String> initialConfig);

    void deploymentDeleted(String deployment);
}
