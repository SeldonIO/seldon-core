package io.seldon.clustermanager.example;

import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;

public class KubernetesManagerNamespacesExample {

    private final static Logger logger = LoggerFactory.getLogger(KubernetesManagerNamespacesExample.class);

    public static class AppConfigKubernetes {

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public KubernetesManager kubernetesManager() {
            return new KubernetesManagerImpl();
        }

    }

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfigKubernetes.class).web(false).run(args);

        KubernetesManager kubernetesManager = ctx.getBean(KubernetesManager.class);
        List<String> namespace_list = kubernetesManager.getNamespaceList();
        System.out.println("Getting namespaces from KubernetesManager");
        System.out.println(namespace_list);

        ctx.close();
    }
}
