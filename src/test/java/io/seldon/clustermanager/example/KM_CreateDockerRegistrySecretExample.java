package io.seldon.clustermanager.example;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef.DockerRegistryDetailsDef;

public class KM_CreateDockerRegistrySecretExample {

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

        final String name = System.getProperty("DOCKER_REGISTRY_SECRET_NAME", "my-registry-secret");
        final String url = System.getProperty("DOCKER_REGISTRY_SECRET_URL", "https://reg.example.com:8080");
        final String username = System.getProperty("DOCKER_REGISTRY_SECRET_USERNAME", "auser");
        final String password = System.getProperty("DOCKER_REGISTRY_SECRET_PASSWORD", "apassword");

        //@formatter:off
        DockerRegistrySecretDef dockerRegistrySecretDef = DockerRegistrySecretDef.newBuilder()
                .setName(name)
                .setDockerRegistryDetails(DockerRegistryDetailsDef.newBuilder()
                        .setUrl(url)
                        .setUsername(username)
                        .setPassword(password)
                        .build())
                .build();
        //@formatter:on
        { // just display the input
            try {
                String json = ProtoBufUtils.toJson(dockerRegistrySecretDef);
                System.out.println("-------------------------------------------------------------------------------");
                System.out.println(json);
                System.out.println("-------------------------------------------------------------------------------");
            } catch (InvalidProtocolBufferException e) {
                e.printStackTrace();
            }
        }
        kubernetesManager.createDockerRegistrySecret(dockerRegistrySecretDef);

        ctx.close();
    }

}
