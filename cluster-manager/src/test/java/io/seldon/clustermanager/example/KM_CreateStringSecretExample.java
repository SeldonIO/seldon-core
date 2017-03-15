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
import io.seldon.protos.DeploymentProtos.StringSecretDef;

public class KM_CreateStringSecretExample {

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

        StringSecretDef stringSecretDef = null;
        {
            //@formatter:off
            stringSecretDef = StringSecretDef.newBuilder()
                    .setName("mysecret")
                    .putData("somekey", "somevalue")
                    .clearType()
                    .build();
            //@formatter:on
        }
        { // just display the input
            try {
                String json = ProtoBufUtils.toJson(stringSecretDef);
                System.out.println("-------------------------------------------------------------------------------");
                System.out.println(json);
                System.out.println("-------------------------------------------------------------------------------");
            } catch (InvalidProtocolBufferException e) {
                e.printStackTrace();
            }
        }
        kubernetesManager.createOrReplaceStringSecret(stringSecretDef);

        ctx.close();
    }

}