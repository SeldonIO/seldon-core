package io.seldon.clustermanager.k8s.tasks;

import java.util.concurrent.Future;

/**
 * Handle tasks based on a key.
 * @author clive
 *
 */
public interface K8sTaskScheduler {
	public Future submit(SeldonDeploymentTaskKey key,Runnable task);
}
