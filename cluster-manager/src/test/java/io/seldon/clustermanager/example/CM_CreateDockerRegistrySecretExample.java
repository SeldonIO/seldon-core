package io.seldon.clustermanager.example;

import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.config.AppConfig;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef.DockerRegistryDetailsDef;

public class CM_CreateDockerRegistrySecretExample {

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfig.class).web(false).run(args);

        ClusterManager clusterManager = ctx.getBean(ClusterManager.class);

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

        try {
            String json = ProtoBufUtils.toJson(dockerRegistrySecretDef);
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println(json);
            System.out.println("-------------------------------------------------------------------------------");
            CMResultDef cmResultDef = clusterManager.createOrReplaceDockerRegistrySecret(dockerRegistrySecretDef);
            String result_json = ProtoBufUtils.toJson(cmResultDef);
            System.out.println(result_json);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        ctx.close();
    }

}
