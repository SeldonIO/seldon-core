package io.seldon.clustermanager.dm;

import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.Future;
import java.util.function.BooleanSupplier;

import org.springframework.beans.factory.annotation.Autowired;

import io.seldon.clustermanager.component.AppComponent;
import io.seldon.clustermanager.dm.DeploymentMonitorAsyncTask.TaskResult;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class DeploymentMonitor implements AppComponent {

    private final static Logger logger = LoggerFactory.getLogger(DeploymentMonitor.class);

    private DeploymentMonitorAsyncTask deploymentMonitorAsyncTask;
    private Map<String, Future<TaskResult>> deploymentsBeingProcessed;

    @Override
    public void init() throws Exception {
        logger.info("init");
        deploymentsBeingProcessed = new HashMap<>();
    }

    @Autowired
    public void setDeploymentMonitorAsyncTask(DeploymentMonitorAsyncTask deploymentMonitorAsyncTask) {
        logger.info("injecting DeploymentMonitorAsyncTask");
        this.deploymentMonitorAsyncTask = deploymentMonitorAsyncTask;
    }

    public void run() {
        System.out.println("** DeploymentMonitor run() **");

        while (true) {
            processAllDeployments();

            try {
                Thread.sleep(100);
            } catch (InterruptedException e) {
                e.printStackTrace();
            }
        }
    }

    @Override
    public void cleanup() throws Exception {
        logger.info("cleanup");
    }

    private void processAllDeployments() {
        List<String> seldonDeploymentIds = Arrays.asList("d0", "d1");

        for (String seldonDeploymentId : seldonDeploymentIds) {
            processDeployment(seldonDeploymentId);
        }
    }

    private void processDeployment(String seldonDeploymentId) {

        BooleanSupplier taskCanRun = () -> {
            Future<TaskResult> futureResult = deploymentsBeingProcessed.get(seldonDeploymentId);
            if (futureResult == null) {
                return true;
            } else if (futureResult.isDone()) {
                deploymentsBeingProcessed.remove(seldonDeploymentId);
                return true;
            }
            return false;
        };

        if (taskCanRun.getAsBoolean()) {
            deploymentsBeingProcessed.put(seldonDeploymentId, deploymentMonitorAsyncTask.runTask(seldonDeploymentId));
        }
    }
}
