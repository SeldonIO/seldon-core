package io.seldon.wrapper;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.condition.ConditionalOnExpression;
import org.springframework.context.annotation.Import;
import org.springframework.scheduling.annotation.EnableAsync;

import io.seldon.wrapper.config.AppConfig;

@SpringBootApplication
@EnableAsync
@Import({ AppConfig.class })
public class App {
    public static void main(String[] args) throws Exception {
        SpringApplication.run(App.class, args);
    }
    
  
}
