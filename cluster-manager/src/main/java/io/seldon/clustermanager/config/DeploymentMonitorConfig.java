package io.seldon.clustermanager.config;

import java.util.concurrent.Executors;

import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Scope;
import org.springframework.core.task.TaskExecutor;
import org.springframework.scheduling.annotation.EnableAsync;
import org.springframework.scheduling.concurrent.ConcurrentTaskExecutor;

import io.seldon.clustermanager.dm.DeploymentMonitor;
import io.seldon.clustermanager.dm.DeploymentMonitorAsyncTask;

@EnableAsync
@Configuration
public class DeploymentMonitorConfig {

    @Bean(initMethod = "init", destroyMethod = "cleanup")
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public DeploymentMonitor deploymentMonitor() {
        return new DeploymentMonitor();
    }

    @Bean
    public DeploymentMonitorAsyncTask deploymentMonitorTask() {
        return new DeploymentMonitorAsyncTask();
    }

    @Bean
    @Qualifier("taskExecutor1")
    public TaskExecutor taskExecutor1() {
        return new ConcurrentTaskExecutor(Executors.newFixedThreadPool(10));
    }

}
