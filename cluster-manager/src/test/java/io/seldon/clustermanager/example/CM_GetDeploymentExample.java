package io.seldon.clustermanager.example;

import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.config.AppConfig;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class CM_GetDeploymentExample {

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfig.class).web(false).run(args);

        ClusterManager clusterManager = ctx.getBean(ClusterManager.class);
        DeploymentDef exampleDeploymentDef = KubernetesManagerExampleUtils.buildExampleDeploymentDef();
        DeploymentDef.Builder deploymentDefBuiler = DeploymentDef.newBuilder();
        deploymentDefBuiler.setId(exampleDeploymentDef.getId());
        DeploymentDef deploymentDefIn = deploymentDefBuiler.build();
        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("deploymentDefIn:");
            String s = ProtoBufUtils.toJson(deploymentDefIn, true);
            System.out.println(s);
            System.out.println("-------------------------------------------------------------------------------");
            CMResultDef cmResultDef = clusterManager.getSeldonDeployment(deploymentDefIn);
            String result_json = ProtoBufUtils.toJson(cmResultDef);
            System.out.println(result_json);
            System.out.println("-------------------------------------------------------------------------------");
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }

        ctx.close();
    }

}
