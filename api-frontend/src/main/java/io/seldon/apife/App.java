package io.seldon.apife;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableAsync;

import com.ryantenney.metrics.spring.config.annotation.EnableMetrics;
import com.ryantenney.metrics.spring.config.annotation.MetricsConfigurerAdapter;

@SpringBootApplication
@EnableAsync
@EnableMetrics(proxyTargetClass = true)
public class App extends MetricsConfigurerAdapter {
    public static void main(String[] args) throws Exception {
        SpringApplication.run(App.class, args);
    }
}
