package io.seldon.clustermanager.example;

import org.springframework.beans.BeansException;
import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class KM_DeleteDeploymentExample {

    public static class ApplicationContextProvider implements ApplicationContextAware {

        private static ApplicationContext _applicationContext;

        @Override
        public void setApplicationContext(ApplicationContext applicationContext) throws BeansException {
            _applicationContext = applicationContext;
        }

        public static ApplicationContext getContext() {
            return _applicationContext;
        }

    }

    public static class AppConfigKubernetes {

        @Bean
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public ApplicationContextProvider applicationContextProvider() {
            return new ApplicationContextProvider();
        }

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public KubernetesManager kubernetesManager() {
            return new KubernetesManagerImpl();
        }

    }

    public static void main(String[] args) {
        new SpringApplicationBuilder(AppConfigKubernetes.class).web(false).run(args);

        ApplicationContext ctx = ApplicationContextProvider.getContext();
        KubernetesManager kubernetesManager = ctx.getBean(KubernetesManager.class);
        DeploymentDef exampleDeploymentDef = KubernetesManagerExampleUtils.buildExampleDeploymentDef();
        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("exampleDeploymentDef:");
            String s = ProtoBufUtils.toJson(exampleDeploymentDef, true);
            System.out.println(s);
            System.out.println("-------------------------------------------------------------------------------");
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        kubernetesManager.deleteSeldonDeployment(exampleDeploymentDef);
    }

}