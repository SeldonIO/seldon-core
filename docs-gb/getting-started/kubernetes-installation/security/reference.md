# Security Settings Reference

## Helm Settings

```yaml
# k8s/helm-charts/seldon-core-v2-setup/values.yaml
security:
  controlplane:
    protocol: PLAINTEXT
    ssl:
      server:
        secret: seldon-controlplane-server
        clientValidationSecret: seldon-controlplane-client
        keyPath: /tmp/certs/cps/tls.key
        crtPath: /tmp/certs/cps/tls.crt
        caPath: /tmp/certs/cps/ca.crt
        clientCaPath: /tmp/certs/cpc/ca.crt
      client:
        secret: seldon-controlplane-client
        serverValidationSecret: seldon-controlplane-server
        keyPath: /tmp/certs/cpc/tls.key
        crtPath: /tmp/certs/cpc/tls.crt
        caPath: /tmp/certs/cpc/ca.crt
        serverCaPath: /tmp/certs/cps/ca.crt
  kafka:
    protocol: PLAINTEXT
    sasl:
      mechanism: SCRAM-SHA-512
      client:
        username: seldon
        secret:
        passwordPath: password
    ssl:
      client:
        secret:
        brokerValidationSecret:
        keyPath: /tmp/certs/kafka/client/tls.key
        crtPath: /tmp/certs/kafka/client/tls.crt
        caPath: /tmp/certs/kafka/client/ca.crt
        brokerCaPath: /tmp/certs/kafka/broker/ca.crt
        endpointIdentificationAlgorithm:
  envoy:
    protocol: PLAINTEXT
    ssl:
      upstream:
        server:
          secret: seldon-upstream-server
          clientValidationSecret: seldon-upstream-client
          keyPath: /tmp/certs/dus/tls.key
          crtPath: /tmp/certs/dus/tls.crt
          caPath: /tmp/certs/dus/ca.crt
          clientCaPath: /tmp/certs/duc/ca.crt
        client:
          secret: seldon-upstream-client
          serverValidationSecret: seldon-upstream-server
          keyPath: /tmp/certs/duc/tls.key
          crtPath: /tmp/certs/duc/tls.crt
          caPath: /tmp/certs/duc/ca.crt
          serverCaPath: /tmp/certs/dus/ca.crt
      downstream:
        server:
          secret: seldon-downstream-server
          clientValidationSecret:
          keyPath: /tmp/certs/dds/tls.key
          crtPath: /tmp/certs/dds/tls.crt
          caPath: /tmp/certs/dds/ca.crt
          clientCaPath: /tmp/certs/ddc/ca.crt
        client:
          mtls: false
          secret:
          serverValidationSecret: seldon-downstream-server
          keyPath: /tmp/certs/ddc/tls.key
          crtPath: /tmp/certs/ddc/tls.crt
          caPath: /tmp/certs/ddc/ca.crt
          serverCaPath: /tmp/certs/dds/ca.crt

# A list of image pull secrets
imagePullSecrets:
```

## Environment variables

Kubernetes secrets and mounted files can be used to provide the certificates in PEM format. These are controlled by environment variables for server or client depending on the component:

### Control Plane

| EnvVar | Value |
| --- | --- |
| CONTROL_PLANE_SECURITY_PROTOCOL | SSL or PLAINTEXT |

For a server (scheduler):

| EnvVar | Value |
| --- | --- |
| CONTROL_PLANE_SERVER_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the certificates |
| CONTROL_PLANE_CLIENT_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the validation ca roots to verify the client certificate |
| CONTROL_PLANE_SERVER_TLS_KEY_LOCATION | the path to the TLS private key |
| CONTROL_PLANE_SERVER_TLS_CRT_LOCATION | the path to the TLS certificate |
| CONTROL_PLANE_SERVER_TLS_CA_LOCATION | the path to the TLS CA chain for the server |
| CONTROL_PLANE_CLIENT_TLS_CA_LOCATION | the path to the TLS CA chain for the client for mTLS verification |

For a client (agent, modelgateway, hodometer, CRD controller):


| EnvVar | Value |
| --- | --- |
| CONTROL_PLANE_CLIENT_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the certificates |
| CONTROL_PLANE_SERVER_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the validation ca roots to verify the server certificate |
| CONTROL_PLANE_CLIENT_TLS_KEY_LOCATION | the path to the TLS private key |
| CONTROL_PLANE_CLIENT_TLS_CRT_LOCATION | the path to the TLS certificate |
| CONTROL_PLANE_CLIENT_TLS_CA_LOCATION | the path to the TLS CA chain for the client |
| CONTROL_PLANE_SERVER_TLS_CA_LOCATION | the path to the TLS CA chain for the server for mTLS verification |


### Kafka

| EnvVar | Value |
| --- | --- |
| KAFKA_SECURITY_PROTOCOL | PLAINTXT, SSL, or SASL_SSL |

| EnvVar | Value |
| --- | --- |
| KAFKA_CLIENT_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the Kafka client certificate |
| KAFKA_CLIENT_SERVER_TLS_KEY_LOCATION | the path to the TLS private key |
| KAFKA_CLIENT_SERVER_TLS_CRT_LOCATION | the path to the TLS certificate |
| KAFKA_CLIENT_SERVER_TLS_CA_LOCATION | the path to the CA chain for the client |
| KAFKA_BROKER_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the validation ca roots for the kafka broker |
| KAFKA_BROKER_TLS_CA_LOCATION | The path to the broker validatiob CA chain |
| KAFKA_CLIENT_SASL_USERNAME | SASL username |
| KAFKA_CLIENT_SASL_SECRET_NAME | the name of the namespaced secret which holds the SASL password |
| KAFKA_CLIENT_SASL_PASSWORD_LOCATION | the path to the file containing the SASL password |


### Envoy

Envoy xDS server will use the control plane server and client certificates defined above.

Downstream server

| EnvVar | Value |
| --- | --- |
| ENVOY_DOWNSTREAM_SERVER_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the certificates |
| ENVOY_DOWNSTREAM_CLIENT_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the validation ca roots to verify the client certificate |
| ENVOY_DOWNSTREAM_SERVER_TLS_KEY_LOCATION | the path to the TLS private key |
| ENVOY_DOWNSTREAM_SERVER_TLS_CRT_LOCATION | the path to the TLS certificate |
| ENVOY_DOWNSTREAM_SERVER_TLS_CA_LOCATION | the path to the TLS CA chain for the server |
| ENVOY_DOWNSTREAM_CLIENT_TLS_CA_LOCATION | the path to the TLS CA chain for the client for mTLS verification |


Downstream client

| EnvVar | Value |
| --- | --- |
| ENVOY_DOWNSTREAM_CLIENT_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the certificates |
| ENVOY_DOWNSTREAM_SERVER_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the validation ca roots to verify the server certificate |
| ENVOY_DOWNSTREAM_CLIENT_TLS_KEY_LOCATION | the path to the TLS private key |
| ENVOY_DOWNSTREAM_CLIENT_TLS_CRT_LOCATION | the path to the TLS certificate |
| ENVOY_DOWNSTREAM_CLIENT_TLS_CA_LOCATION | the path to the TLS CA chain for the server |
| ENVOY_DOWNSTREAM_SERVER_TLS_CA_LOCATION | the path to the TLS CA chain for the server for mTLS verification |


Upstream server

| EnvVar | Value |
| --- | --- |
| ENVOY_UPSTREAM_SERVER_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the certificates |
| ENVOY_UPSTREAM_CLIENT_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the validation ca roots to verify the client certificate |
| ENVOY_UPSTREAM_SERVER_TLS_KEY_LOCATION | the path to the TLS private key |
| ENVOY_UPSTREAM_SERVER_TLS_CRT_LOCATION | the path to the TLS certificate |
| ENVOY_UPSTREAM_SERVER_TLS_CA_LOCATION | the path to the TLS CA chain for the server |
| ENVOY_UPSTREAM_CLIENT_TLS_CA_LOCATION | the path to the TLS CA chain for the client for mTLS verification |


Upstream client

| EnvVar | Value |
| --- | --- |
| ENVOY_UPSTREAM_CLIENT_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the certificates |
| ENVOY_UPSTREAM_SERVER_TLS_SECRET_NAME | (optional) the name of the namespaced secret which holds the validation ca roots to verify the server certificate |
| ENVOY_UPSTREAM_CLIENT_TLS_KEY_LOCATION | the path to the TLS private key |
| ENVOY_UPSTREAM_CLIENT_TLS_CRT_LOCATION | the path to the TLS certificate |
| ENVOY_UPSTREAM_CLIENT_TLS_CA_LOCATION | the path to the TLS CA chain for the server |
| ENVOY_UPSTREAM_SERVER_TLS_CA_LOCATION | the path to the TLS CA chain for the server for mTLS verificatio |n
