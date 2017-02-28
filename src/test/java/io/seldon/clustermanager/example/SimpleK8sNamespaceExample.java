package io.seldon.clustermanager.example;

import org.springframework.beans.factory.config.ConfigurableBeanFactory;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Scope;

import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.util.JsonFormat;

import io.seldon.clustermanager.component.KubernetesManager;
import io.seldon.clustermanager.k8s.KubernetesManagerImpl;
import io.seldon.protos.DeploymentProtos.CMResultDef;

public class SimpleK8sNamespaceExample {

    public static class AppConfigKubernetes {

        @Bean(initMethod = "init", destroyMethod = "cleanup")
        @Scope(ConfigurableBeanFactory.SCOPE_SINGLETON)
        public KubernetesManager kubernetesManager() {
            return new KubernetesManagerImpl();
        }

    }

    public static void main(String[] args) {
        AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext();
        ctx.register(AppConfigKubernetes.class);
        ctx.refresh();

        KubernetesManager kubernetesManager = ctx.getBean(KubernetesManager.class);
        CMResultDef cmResultDef = ((KubernetesManagerImpl) kubernetesManager).getNamespaces();
        String json = null;
        try {
            json = toJson(cmResultDef);
        } catch (InvalidProtocolBufferException e) {
            e.printStackTrace();
        }
        System.out.println(json);

        ctx.close();
    }

    private static String toJson(Message message) throws InvalidProtocolBufferException {
        String json = null;
        json = JsonFormat.printer().includingDefaultValueFields().preservingProtoFieldNames().print(message);
        return json;
    }
}
