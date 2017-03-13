package io.seldon.clustermanager.example;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;

public class KM_DeleteDockerRegistrySecretExample {

    public static class AppConfigKubernetes {

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public KubernetesManager kubernetesManager() {
            return new KubernetesManagerImpl();
        }

    }

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfigKubernetes.class).web(false).run(args);

        final String name = System.getProperty("DOCKER_REGISTRY_SECRET_NAME", "my-registry-secret");
        System.out.println("-------------------------------------------------------------------------------");
        System.out.println("Trying to delete: "+name);
        System.out.println("-------------------------------------------------------------------------------");

        KubernetesManager kubernetesManager = ctx.getBean(KubernetesManager.class);
        kubernetesManager.deleteDockerRegistrySecret(name);
        ctx.close();
    }

}