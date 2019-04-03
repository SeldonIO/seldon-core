package io.seldon.clustermanager.k8s.tasks;

import java.util.ArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * Manage tasks by ensuring they are executed in series for for each SeldonDeployment
 * @author clive
 *
 */
@Component
public class K8sTaskSchedulerImpl implements K8sTaskScheduler {
		
	protected static Logger logger = LoggerFactory.getLogger(K8sTaskSchedulerImpl.class.getName());
	private static final int NUM_EXXECUTORS = 20;
	
	ArrayList<ExecutorService> executors;
	
	/**
	 * Setup Executors.
	 */
	public K8sTaskSchedulerImpl()
	{
		executors = new ArrayList<>();
		for(int i =0;i<NUM_EXXECUTORS;i++)
			executors.add(Executors.newFixedThreadPool(1));	
	}
	
	/**
	 * Submit a task to run. Choose one of the executors to run in based on hash of key.
	 * @param task
	 */
	public Future<?> submit(SeldonDeploymentTaskKey key,Runnable task)
	{
		//Use floorMod as hash can be negative
		final int idx = Math.floorMod (key.hashCode(), NUM_EXXECUTORS);
		logger.info("Adding task to executor {}",idx);
		return executors.get(idx).submit(task);
	}

}
