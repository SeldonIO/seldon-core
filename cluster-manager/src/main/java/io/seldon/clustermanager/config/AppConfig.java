package io.seldon.clustermanager.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;

import io.seldon.clustermanager.ClusterManagerProperites;

public class AppConfig {
    

    @ConfigurationProperties(prefix = "io.seldon.clustermanager")
    @Bean
    public ClusterManagerProperites clusterManagerProperites() {
        return new ClusterManagerProperites();
    }
}
