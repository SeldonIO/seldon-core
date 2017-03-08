package io.seldon.clustermanager.example;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.boot.builder.SpringApplicationBuilder;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import io.seldon.clustermanager.component.ZookeeperManager;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.clustermanager.zk.ZookeeperManagerImpl;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class ZK_DeleteDeploymentExample {

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
        try {
            System.out.println("-------------------------------------------------------------------------------");
            System.out.println("exampleDeploymentDef:");
            String s = ProtoBufUtils.toJson(exampleDeploymentDef, true);
            System.out.println(s);
            System.out.println("-------------------------------------------------------------------------------");
            zookeeperManager.deleteDeployment(exampleDeploymentDef);
        } catch (Exception e) {
            e.printStackTrace();
        }

        ctx.close();
    }

}
