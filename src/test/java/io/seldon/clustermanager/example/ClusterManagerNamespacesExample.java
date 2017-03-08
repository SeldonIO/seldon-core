package io.seldon.clustermanager.example;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Scope;

import com.google.protobuf.InvalidProtocolBufferException;

import io.seldon.clustermanager.cm.CluserManagerImpl;
import io.seldon.clustermanager.component.ClusterManager;
import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.component.ZookeeperManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;
import io.seldon.clustermanager.pb.ProtoBufUtils;
import io.seldon.protos.DeploymentProtos.CMResultDef;
import io.seldon.protos.DeploymentProtos.DeploymentDef;

public class ClusterManagerNamespacesExample {

    public static class CustomConfig {

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public ZookeeperManager zookeeperManager() {
            return new ZookeeperManager() {

                @Override
                public void init() throws Exception {
                }

                @Override
                public void cleanup() throws Exception {
                }

                @Override
                public void persistDeployment(DeploymentDef deploymentDef) throws Exception {
                }

                @Override
                public void deleteDeployment(DeploymentDef deploymentDef) throws Exception {
                }
            };
        }

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public KubernetesManager kubernetesManager() {
            return new KubernetesManagerImpl();
        }

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public ClusterManager clusterManager() {
            return new CluserManagerImpl();
        }
    }

    public static void main(String[] args) {
        AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext();
        ctx.register(CustomConfig.class);
        ctx.refresh();

        ClusterManager clusterManager = ctx.getBean(ClusterManager.class);
        CMResultDef cmResultDef = clusterManager.getNamespaces();
        String json = null;
        try {
            json = ProtoBufUtils.toJson(cmResultDef);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
        System.out.println(json);

        ctx.close();
    }
}
