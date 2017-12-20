package io.seldon.apife.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;

import io.seldon.apife.AppProperties;

public class AppConfig {
    

    @ConfigurationProperties(prefix = "io.seldon.apife")
    @Bean
    public AppProperties appProperites() {
        return new AppProperties();
    }
}
