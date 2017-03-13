package io.seldon.clustermanager.example;

import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.config.AppConfig;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;

public class CM_DeleteDockerRegistrySecretExample {

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfig.class).web(false).run(args);

        ClusterManager clusterManager = ctx.getBean(ClusterManager.class);

        final String name = System.getProperty("DOCKER_REGISTRY_SECRET_NAME", "my-registry-secret");

        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("Trying to delete: " + name);
            System.out.println("-------------------------------------------------------------------------------");
            CMResultDef cmResultDef = clusterManager.deleteDockerRegistrySecret(name);
            String result_json = ProtoBufUtils.toJson(cmResultDef);
            System.out.println(result_json);
            System.out.println("-------------------------------------------------------------------------------");
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        ctx.close();
    }

}