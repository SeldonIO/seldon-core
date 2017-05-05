package io.seldon.clustermanager.dm;

import java.util.concurrent.Future;

import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.AsyncResult;

public class DeploymentMonitorAsyncTask {

    public class TaskResult {
    }

    @Async("taskExecutor1")
    public Future<TaskResult> runTask(String seldonDeploymentId) {
        System.out.println(seldonDeploymentId + " is running and about to SLEEP");
        System.out.println(Thread.currentThread().getName());

        try {
            Thread.sleep(5000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }

        System.out.println(seldonDeploymentId + " is running and about to FINISH");

        return new AsyncResult<>(new TaskResult());
    }
}
