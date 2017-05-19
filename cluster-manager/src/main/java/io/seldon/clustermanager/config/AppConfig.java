package io.seldon.clustermanager.config;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Scope;

import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.cm.CluserManagerImpl;
import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.component.ZookeeperManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;
import io.seldon.clustermanager.zk.ZookeeperManagerImpl;

public class AppConfig {
    @Bean(initMethod = "init", destroyMethod = "cleanup")
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public ZookeeperManager zookeeperManager() {
        return new ZookeeperManagerImpl();
    }
    
    @Bean(initMethod = "init", destroyMethod = "cleanup")
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public KubernetesManager kubernetesManager() {
        return new KubernetesManagerImpl();
    }

    @Bean(initMethod = "init", destroyMethod = "cleanup")
    @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
    public ClusterManager clusterManager() {
        return new CluserManagerImpl();
    }
    
    @ConfigurationProperties(prefix = "io.seldon.clustermanager")
    @Bean
    public ClusterManagerProperites clusterManagerProperites() {
        return new ClusterManagerProperites();
    }
}
