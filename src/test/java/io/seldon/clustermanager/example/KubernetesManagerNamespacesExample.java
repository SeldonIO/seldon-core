package io.seldon.clustermanager.example;

import java.util.List;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;

public class KubernetesManagerNamespacesExample {

    public static class AppConfigKubernetes {

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public KubernetesManager kubernetesManager() {
            return new KubernetesManagerImpl();
        }

    }

    public static void main(String[] args) {
        AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext();
        ctx.register(AppConfigKubernetes.class);
        ctx.refresh();

        KubernetesManager kubernetesManager = ctx.getBean(KubernetesManager.class);
        List<String> namespace_list = kubernetesManager.getNamespaceList();
        System.out.println("Getting namespaces from KubernetesManager");
        System.out.println(namespace_list);

        ctx.close();
    }
}
