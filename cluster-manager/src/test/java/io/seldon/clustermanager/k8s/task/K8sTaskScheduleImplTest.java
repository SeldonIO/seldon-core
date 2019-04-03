package io.seldon.clustermanager.k8s.task;

import java.util.concurrent.ExecutionException;
import java.util.concurrent.Future;

import org.junit.Test;

import io.seldon.clustermanager.k8s.tasks.K8sTaskSchedulerImpl;
import io.seldon.clustermanager.k8s.tasks.SeldonDeploymentTaskKey;
import org.junit.Assert;

public class K8sTaskScheduleImplTest {
	
	public static class ConstantTask implements Runnable {

		final static int WAIT_TIME = 500;
		long startTime;
		long endTime;
		
		@Override
		public void run() {
			this.startTime = System.currentTimeMillis();
			try {
				Thread.sleep(WAIT_TIME);
			} catch (InterruptedException e) {
				// TODO Auto-generated catch block
				e.printStackTrace();
			}
			this.endTime = System.currentTimeMillis();
		}
		
	}

	
	@Test
	public void basicTest() throws InterruptedException, ExecutionException
	{
		K8sTaskSchedulerImpl scheduler = new K8sTaskSchedulerImpl();
		
		SeldonDeploymentTaskKey key = new SeldonDeploymentTaskKey("a", "b", "c");
		ConstantTask task = new ConstantTask();
		
		long start = System.currentTimeMillis();
		Future<?> f = scheduler.submit(key, task);
		f.get();
		long end = System.currentTimeMillis();
		long diff = end-start;
		long diff2 = task.endTime - start;
		Assert.assertTrue(diff2 >= ConstantTask.WAIT_TIME);
		Assert.assertTrue(diff >= diff2);
		
	}
	
	
	@Test
	public void testParallel() throws InterruptedException, ExecutionException
	{
		K8sTaskSchedulerImpl scheduler = new K8sTaskSchedulerImpl();
		
		SeldonDeploymentTaskKey key1 = new SeldonDeploymentTaskKey("a", "b", "c1");
		SeldonDeploymentTaskKey key2 = new SeldonDeploymentTaskKey("a", "b", "c2");
		ConstantTask task1 = new ConstantTask();
		ConstantTask task2 = new ConstantTask();
		
		long start = System.currentTimeMillis();
		Future<?> f1 = scheduler.submit(key1, task1);
		Future<?> f2 = scheduler.submit(key2, task2);
		f1.get();
		f2.get();
		long end = System.currentTimeMillis();
		long diff = end-start;
		long diff1 = task1.endTime - start;
		long diff2 = task2.endTime - start;
		Assert.assertTrue(task2.startTime < task1.endTime);
		Assert.assertTrue(diff1 >= ConstantTask.WAIT_TIME);
		Assert.assertTrue(diff2 >= ConstantTask.WAIT_TIME);

	}
	
	@Test
	public void testSequential() throws InterruptedException, ExecutionException
	{
		K8sTaskSchedulerImpl scheduler = new K8sTaskSchedulerImpl();
		
		SeldonDeploymentTaskKey key1 = new SeldonDeploymentTaskKey("a", "b", "c");
		SeldonDeploymentTaskKey key2 = new SeldonDeploymentTaskKey("a", "b", "c");
		ConstantTask task1 = new ConstantTask();
		ConstantTask task2 = new ConstantTask();
		
		long start = System.currentTimeMillis();
		Future<?> f1 = scheduler.submit(key1, task1);
		Future<?> f2 = scheduler.submit(key2, task2);
		f1.get();
		f2.get();
		long end = System.currentTimeMillis();
		long diff = end-start;
		long diff1 = task1.endTime - start;
		long diff2 = task2.endTime - start;
		Assert.assertFalse(task2.startTime < task1.endTime);
		Assert.assertTrue(diff1 >= ConstantTask.WAIT_TIME);
		Assert.assertTrue(diff2 >= ConstantTask.WAIT_TIME);

	}
	
}
