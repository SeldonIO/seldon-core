package io.seldon.example.h2o;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Import;

@SpringBootApplication(scanBasePackages = {"io.seldon.wrapper","io.seldon.example.h2o"})
@Import({ io.seldon.wrapper.config.AppConfig.class })
public class App {
    public static void main(String[] args) throws Exception {
        SpringApplication.run(App.class, args);
    }
}
