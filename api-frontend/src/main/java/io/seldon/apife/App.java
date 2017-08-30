package io.seldon.apife;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableAsync;

import io.micrometer.spring.export.prometheus.EnablePrometheusMetrics;

@SpringBootApplication
@EnableAsync
@EnablePrometheusMetrics
public class App  {
    public static void main(String[] args) throws Exception {
        SpringApplication.run(App.class, args);
    }
}
