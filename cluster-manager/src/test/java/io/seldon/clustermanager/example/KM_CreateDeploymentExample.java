package io.seldon.clustermanager.example;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class KM_CreateDeploymentExample {

    @EnableAutoConfiguration
    public static class AppConfigKubernetes {

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

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfigKubernetes.class).web(false).run(args);

        KubernetesManager kubernetesManager = ctx.getBean(KubernetesManager.class);
        DeploymentDef exampleDeploymentDef = KubernetesManagerExampleUtils.buildExampleDeploymentDef();
        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("exampleDeploymentDef:");
            String s = ProtoBufUtils.toJson(exampleDeploymentDef, true);
            System.out.println(s);
            System.out.println("-------------------------------------------------------------------------------");
            DeploymentDef resultingDeploymentDef = kubernetesManager.createOrReplaceSeldonDeployment(exampleDeploymentDef,null);
            String r = ProtoBufUtils.toJson(resultingDeploymentDef, true);
            System.out.println(r);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        ctx.close();
    }

}
