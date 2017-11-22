package io.seldon.clustermanager.config;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;

public class AppConfig {
    
    @Bean(initMethod = "init", destroyMethod = "cleanup")
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public KubernetesManager kubernetesManager() {
        return new KubernetesManagerImpl();
    }

    @ConfigurationProperties(prefix = "io.seldon.clustermanager")
    @Bean
    public ClusterManagerProperites clusterManagerProperites() {
        return new ClusterManagerProperites();
    }
}
