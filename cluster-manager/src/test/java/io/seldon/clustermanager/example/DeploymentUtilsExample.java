package io.seldon.clustermanager.example;

import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.ClusterManagerProperites;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.DeploymentDef;


public class DeploymentUtilsExample {

    @EnableAutoConfiguration
    public static class AppConfig {

        @ConfigurationProperties(prefix = "io.seldon.clustermanager")
        @Bean
        public ClusterManagerProperites clusterManagerProperites() {
            return new ClusterManagerProperites();
        }
    }

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfig.class).web(false).run(args);

        DeploymentDef exampleDeploymentDef = KubernetesManagerExampleUtils.buildExampleDeploymentDef();

        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("exampleDeploymentDef:");
            String s = ProtoBufUtils.toJson(exampleDeploymentDef, true,false);
            System.out.println(s);
            System.out.println("-------------------------------------------------------------------------------");
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        ClusterManagerProperites clusterManagerProperites = ctx.getBean(ClusterManagerProperites.class);
        //FIXME
        //List<BuildDeploymentResult> buildDeploymentResults = DeploymentUtils.buildDeployments(exampleDeploymentDef);
        //System.out.println(buildDeploymentResults);

        ctx.close();
    }
}
