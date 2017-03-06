package io.seldon.clustermanager.k8s;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.fabric8.kubernetes.api.model.Service;
import io.fabric8.kubernetes.api.model.ServiceBuilder;
import io.fabric8.kubernetes.api.model.extensions.Deployment;
import io.fabric8.kubernetes.client.KubernetesClient;

public class KubernetesServiceOps {

    private final static Logger logger = LoggerFactory.getLogger(KubernetesServiceOps.class);

    private final KubernetesClient kubernetesClient;
    private final String namespace_name;
    private final Deployment deployment;

    public KubernetesServiceOps(KubernetesClient kubernetesClient, String namespace_name, Deployment deployment) {
        this.kubernetesClient = kubernetesClient;
        this.namespace_name = namespace_name;
        this.deployment = deployment;
    }

    public Service create() {
        final String deploymentName = deployment.getMetadata().getName();
        String serviceName = deploymentName;

        String selectorName = "seldon-app";
        String selectorValue = deploymentName;

        int port = 8000;
        int targetPort = 80;

        //@formatter:off
        Service service = new ServiceBuilder()
                .withNewMetadata()
                    .withName(serviceName)
                .endMetadata()
                .withNewSpec()
                    .addNewPort()
                        .withProtocol("TCP")
                        .withPort(port)
                        .withNewTargetPort(targetPort)
                    .endPort()
                    .addToSelector(selectorName, selectorValue)
                    .withType("ClusterIP")
                .endSpec()
                .build();
        //@formatter:on

        service = kubernetesClient.services().inNamespace(namespace_name).create(service);
        String serviceNameForMsg = (service != null) ? service.getMetadata().getName() : null;
        logger.debug(String.format("Created kubernetes service [%s]", serviceNameForMsg));
        return service;
    }

    public Service update() {
        final String deploymentName = deployment.getMetadata().getName();
        String serviceName = deploymentName;

        String selectorName = "seldon-app";
        String selectorValue = deploymentName;

        int port = 8000;
        int targetPort = 80;

        //@formatter:off
        Service service = new ServiceBuilder()
                .withNewMetadata()
                    .withName(serviceName)
                .endMetadata()
                .withNewSpec()
                    .addNewPort()
                        .withProtocol("TCP")
                        .withPort(port)
                        .withNewTargetPort(targetPort)
                    .endPort()
                    .addToSelector(selectorName, selectorValue)
                    .withType("ClusterIP")
                .endSpec()
                .build();
        //@formatter:on

        service = kubernetesClient.services().inNamespace(namespace_name).createOrReplace(service);
        String serviceNameForMsg = (service != null) ? service.getMetadata().getName() : null;
        logger.debug(String.format("Updated kubernetes service [%s]", serviceNameForMsg));
        return service;
    }

    public void delete() {
        final String deploymentName = deployment.getMetadata().getName();
        String serviceName = deploymentName;

        Service service = kubernetesClient.services().inNamespace(namespace_name).withName(serviceName).get();
        kubernetesClient.resource(service).delete();
        logger.debug(String.format("Deleted kubernetes service [%s]", service.getMetadata().getName()));
    }
}
