package io.seldon.clustermanager.example;

import java.util.List;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.SpringBootConfiguration;
import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.PropertySource;
import org.springframework.context.annotation.Scope;
import org.springframework.context.support.PropertySourcesPlaceholderConfigurer;

import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class KubernetesManagerNamespacesExample {

    private final  static Logger logger = LoggerFactory.getLogger(KubernetesManagerNamespacesExample.class);
    
    //@PropertySource("classpath:/application.properties")
    
    @PropertySource("classpath:/application.properties")
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

        logger.info("OK!");
        KubernetesManager kubernetesManager = ctx.getBean(KubernetesManager.class);
        List<String> namespace_list = kubernetesManager.getNamespaceList();
        System.out.println("Getting namespaces from KubernetesManager");
        System.out.println(namespace_list);

        ctx.close();
    }
}
