package io.seldon.clustermanager.example;

import java.util.Optional;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import io.seldon.clustermanager.component.ZookeeperManager;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.clustermanager.zk.ZookeeperManagerImpl;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class ZK_GetDeploymentExample {

    public static class AppConfig {
        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public ZookeeperManager zookeeperManager() {
            return new ZookeeperManagerImpl();
        }
    }

    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = new SpringApplicationBuilder(AppConfig.class).web(false).run(args);

        ZookeeperManager zookeeperManager = ctx.getBean(ZookeeperManager.class);

        DeploymentDef exampleDeploymentDef = KubernetesManagerExampleUtils.buildExampleDeploymentDef();
        DeploymentDef.Builder deploymentDefBuiler = DeploymentDef.newBuilder();
        deploymentDefBuiler.setId(exampleDeploymentDef.getId());
        DeploymentDef deploymentDefIn = deploymentDefBuiler.build();
        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("input deploymentDef:");
            String sin = ProtoBufUtils.toJson(deploymentDefIn, true);
            System.out.println(sin);
            System.out.println("-------------------------------------------------------------------------------");
            Optional<DeploymentDef> deploymentDefOut = zookeeperManager.getSeldonDeployment(deploymentDefIn);
            if (deploymentDefOut.isPresent()) {
                System.out.println("output deploymentDef:");
                String sout = ProtoBufUtils.toJson(deploymentDefOut.get(), true);
                System.out.println(sout);
            } else {
                System.out.println("NOT FOUND!!");
            }
            System.out.println("-------------------------------------------------------------------------------");
        } catch (Exception e) {
            e.printStackTrace();
        }
        ctx.close();
    }

}