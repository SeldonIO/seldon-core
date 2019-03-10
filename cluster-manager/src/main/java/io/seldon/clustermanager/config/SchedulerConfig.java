package io.seldon.clustermanager.config;

import java.util.concurrent.Executor;
import java.util.concurrent.Executors;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.EnableScheduling;

/**
 * Provide multi threaded default scheduling for Spring schedulers
 * @author clive
 *
 */
@Configuration
@EnableScheduling
public class SchedulerConfig {

	@Bean(destroyMethod = "shutdown")
    public Executor taskScheduler() {
        return Executors.newScheduledThreadPool(2);
    }
}
