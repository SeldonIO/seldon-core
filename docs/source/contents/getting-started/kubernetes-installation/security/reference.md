# Security Settings Reference

## Helm Settings

```{literalinclude} ../../../../../../k8s/helm-charts/seldon-core-v2-setup/values.yaml
:language: yaml
   :language: golang
   :start-after: # Security settings
   :end-before: opentelemetry
```


## Environment variables

Kubernetes secrets and mounted files can be used to provide the certificates in PEM format. These are controlled by environment variables for server or client depending on the component:

### Control Plane

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - CONTROL_PLANE_SECURITY_PROTOCOL
  - SSL or PLAINTEXT
```

For a server (scheduler):

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - CONTROL_PLANE_SERVER_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the certificates
* - CONTROL_PLANE_CLIENT_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the validation ca roots to verify the client certificate
* - CONTROL_PLANE_SERVER_TLS_KEY_LOCATION
  - the path to the TLS private key
* - CONTROL_PLANE_SERVER_TLS_CRT_LOCATION
  - the path to the TLS certificate
* - CONTROL_PLANE_SERVER_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server
* - CONTROL_PLANE_CLIENT_TLS_CA_LOCATION
  - the path to the TLS CA chain for the client for mTLS verification
```

For a client (agent, modelgateway, hodometer, CRD controller):


```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - CONTROL_PLANE_CLIENT_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the certificates
* - CONTROL_PLANE_SERVER_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the validation ca roots to verify the server certificate
* - CONTROL_PLANE_CLIENT_TLS_KEY_LOCATION
  - the path to the TLS private key
* - CONTROL_PLANE_CLIENT_TLS_CRT_LOCATION
  - the path to the TLS certificate
* - CONTROL_PLANE_CLIENT_TLS_CA_LOCATION
  - the path to the TLS CA chain for the client
* - CONTROL_PLANE_SERVER_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server for mTLS verification
```

### Kafka

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - KAFKA_SECURITY_PROTOCOL
  - PLAINTXT, SSL, or SASL_SSL
```

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - KAFKA_CLIENT_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the Kafka client certificate
* - KAFKA_CLIENT_SERVER_TLS_KEY_LOCATION
  - the path to the TLS private key
* - KAFKA_CLIENT_SERVER_TLS_CRT_LOCATION
  - the path to the TLS certificate
* - KAFKA_CLIENT_SERVER_TLS_CA_LOCATION
  - the path to the CA chain for the client
* - KAFKA_BROKER_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the validation ca roots for the kafka broker
* - KAFKA_BROKER_TLS_CA_LOCATION
  - The path to the broker validatiob CA chain
* - KAFKA_CLIENT_SASL_USERNAME
  - SASL username
* - KAFKA_CLIENT_SASL_SECRET_NAME
  - the name of the namespaced secret which holds the SASL password
* - KAFKA_CLIENT_SASL_PASSWORD_LOCATION
  - the path to the file containing the SASL password
```

### Envoy

Envoy xDS server will use the control plane server and client certificates defined above.

Downstream server

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - ENVOY_DOWNSTREAM_SERVER_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the certificates
* - ENVOY_DOWNSTREAM_CLIENT_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the validation ca roots to verify the client certificate
* - ENVOY_DOWNSTREAM_SERVER_TLS_KEY_LOCATION
  - the path to the TLS private key
* - ENVOY_DOWNSTREAM_SERVER_TLS_CRT_LOCATION
  - the path to the TLS certificate
* - ENVOY_DOWNSTREAM_SERVER_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server
* - ENVOY_DOWNSTREAM_CLIENT_TLS_CA_LOCATION
  - the path to the TLS CA chain for the client for mTLS verification
```

Downstream client

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - ENVOY_DOWNSTREAM_CLIENT_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the certificates
* - ENVOY_DOWNSTREAM_SERVER_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the validation ca roots to verify the server certificate
* - ENVOY_DOWNSTREAM_CLIENT_TLS_KEY_LOCATION
  - the path to the TLS private key
* - ENVOY_DOWNSTREAM_CLIENT_TLS_CRT_LOCATION
  - the path to the TLS certificate
* - ENVOY_DOWNSTREAM_CLIENT_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server
* - ENVOY_DOWNSTREAM_SERVER_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server for mTLS verification
```

Upstream server

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - ENVOY_UPSTREAM_SERVER_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the certificates
* - ENVOY_UPSTREAM_CLIENT_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the validation ca roots to verify the client certificate
* - ENVOY_UPSTREAM_SERVER_TLS_KEY_LOCATION
  - the path to the TLS private key
* - ENVOY_UPSTREAM_SERVER_TLS_CRT_LOCATION
  - the path to the TLS certificate
* - ENVOY_UPSTREAM_SERVER_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server
* - ENVOY_UPSTREAM_CLIENT_TLS_CA_LOCATION
  - the path to the TLS CA chain for the client for mTLS verification
```

Upstream client

```{list-table}
:header-rows: 1

* - EnvVar
  - Value
* - ENVOY_UPSTREAM_CLIENT_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the certificates
* - ENVOY_UPSTREAM_SERVER_TLS_SECRET_NAME
  - (optional) the name of the namespaced secret which holds the validation ca roots to verify the server certificate
* - ENVOY_UPSTREAM_CLIENT_TLS_KEY_LOCATION
  - the path to the TLS private key
* - ENVOY_UPSTREAM_CLIENT_TLS_CRT_LOCATION
  - the path to the TLS certificate
* - ENVOY_UPSTREAM_CLIENT_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server
* - ENVOY_UPSTREAM_SERVER_TLS_CA_LOCATION
  - the path to the TLS CA chain for the server for mTLS verification
```



