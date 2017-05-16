package io.seldon.clustermanager.k8s;

import java.util.Base64;
import java.util.Map;
import java.util.stream.Collectors;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import io.fabric8.kubernetes.api.model.Secret;
import io.fabric8.kubernetes.api.model.SecretBuilder;
import io.fabric8.kubernetes.client.KubernetesClient;
import io.seldon.protos.DeploymentProtos.DockerRegistrySecretDef;
import io.seldon.protos.DeploymentProtos.StringSecretDef;

public class SecretUtils {

    private final static Logger logger = LoggerFactory.getLogger(SecretUtils.class);

    public static Secret createOrReplaceSecret(KubernetesClient kubernetesClient, String namespace_name, StringSecretDef stringSecretDef) {
        final String name = stringSecretDef.getName();
        final Map<String, String> data = stringSecretDef.getDataMap();
        final String type = stringSecretDef.getType();

        // Transform the values of the map to be base64Encoded
        // eg
        // from {somekey=somevalue}
        // to {somekey=c29tZXZhbHVl}
        Map<String, String> dataBase64Encoded = data.entrySet().stream()
                .collect(Collectors.toMap(Map.Entry::getKey, e -> new String(Base64.getEncoder().encode(e.getValue().getBytes()))));

        //@formatter:off
        Secret secret = new SecretBuilder()
                .withNewMetadata()
                    .withName(name)
                .endMetadata()
                .withType(type)
                .withData(dataBase64Encoded)
                .build();
        //@formatter:on

        secret = kubernetesClient.secrets().inNamespace(namespace_name).createOrReplace(secret);
        logger.debug(String.format("Created kubernetes secret [%s]", name));
        return secret;
    }

    public static Secret createOrReplaceSecret(KubernetesClient kubernetesClient, String namespace_name, DockerRegistrySecretDef dockerRegistrySecretDef) {
        final String name = dockerRegistrySecretDef.getName();
        logger.debug(String.format("Creating kubernetes docker registry secret [%s]", name));
        final String username = dockerRegistrySecretDef.getDockerRegistryDetails().getUsername();
        final String psword = dockerRegistrySecretDef.getDockerRegistryDetails().getPassword();

        String url = dockerRegistrySecretDef.getDockerRegistryDetails().getUrl();
        String auth = username + ":" + psword;
        String authBase64Encoded = new String(Base64.getEncoder().encode(auth.getBytes()));

        ObjectMapper mapper = new ObjectMapper();
        ObjectNode dotDockercfgObject = mapper.createObjectNode();
        //@formatter:off
        dotDockercfgObject
            .set(url, mapper.createObjectNode()
                .put("auth", authBase64Encoded));
        //@formatter:on

        String dotDockercfgString = dotDockercfgObject.toString();
        //@formatter:off
        StringSecretDef stringSecretDef = StringSecretDef.newBuilder()
                .setName(name)
                .putData(".dockercfg", dotDockercfgString)
                .setType("kubernetes.io/dockercfg")
                .build();
        //@formatter:on

        return createOrReplaceSecret(kubernetesClient, namespace_name, stringSecretDef);
    }

    public static void deleteSecret(KubernetesClient kubernetesClient, String namespace_name, String name) {
        kubernetesClient.secrets().inNamespace(namespace_name).withName(name).delete();
        logger.debug(String.format("Deleted kubernetes secret [%s]", name));
    }
}
