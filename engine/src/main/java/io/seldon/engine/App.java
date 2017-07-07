package io.seldon.engine;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Import;
import org.springframework.scheduling.annotation.EnableAsync;

import io.seldon.engine.config.AppConfig;

@SpringBootApplication
@EnableAsync
@Import({ AppConfig.class })
public class App {
    public static void main(String[] args) throws Exception {
        SpringApplication.run(App.class, args);
    }

}
