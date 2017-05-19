package io.seldon.clustermanager.example;

import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;

import io.seldon.clustermanager.ClusterManagerProperites;

public class ClusterManagerProperitesExample {

    @EnableAutoConfiguration
    public static class AppConfig {

        @ConfigurationProperties(prefix = "io.seldon.clustermanager")
        @Bean
        public ClusterManagerProperites clusterManagerProperites() {
            return new ClusterManagerProperites();
        }
    }

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfig.class).web(false).run(args);

        ClusterManagerProperites clusterManagerProperites = ctx.getBean(ClusterManagerProperites.class);
        System.out.println(clusterManagerProperites);

        ctx.close();

    }
}
