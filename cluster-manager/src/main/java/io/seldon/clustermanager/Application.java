package io.seldon.clustermanager;

import java.util.concurrent.ExecutorService;
import java.util.function.BooleanSupplier;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Import;
import org.springframework.scheduling.concurrent.ConcurrentTaskExecutor;

import io.seldon.clustermanager.config.AppConfig;
import io.seldon.clustermanager.config.DeploymentMonitorConfig;
import io.seldon.clustermanager.dm.DeploymentMonitor;
import joptsimple.OptionParser;
import joptsimple.OptionSet;

@SpringBootApplication
@Import({ AppConfig.class })
public class Application {

    private static final String DEPLOYMENT_MONITOR_OPTION_NAME = "deployment-monitor";

    public static void main(String[] args) {

        BooleanSupplier isDeploymentMonitor = () -> {
            OptionParser parser = new OptionParser();
            parser.accepts(DEPLOYMENT_MONITOR_OPTION_NAME);
            OptionSet options = parser.parse(args);
            return options.has(DEPLOYMENT_MONITOR_OPTION_NAME);
        };

        if (!isDeploymentMonitor.getAsBoolean()) {
            SpringApplication.run(Application.class, args);
        } else {
            SpringApplication app = new SpringApplication(AppConfig.class, DeploymentMonitorConfig.class);
            app.setWebEnvironment(false);
            ConfigurableApplicationContext ctx = app.run(args);

            DeploymentMonitor deploymentMonitor = ctx.getBean(DeploymentMonitor.class);
            deploymentMonitor.run();

            ConcurrentTaskExecutor exec = (ConcurrentTaskExecutor) ctx.getBean("taskExecutor1");
            ExecutorService es = (ExecutorService) exec.getConcurrentExecutor();
            es.shutdown();

            ctx.close();
        }
    }
}