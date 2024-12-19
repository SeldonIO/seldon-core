# Managed Kafka

Seldon recommends a [managed Kafka](../#managed-kafka) for production installation. On this page we will demonstrate how you can integrate and configure your managed Kafka with Seldon Core 2. 

1. [Examples configurations](managed-kafka.md#example-configurations-for-managed-kafka-services)
2. [Configuring with Seldon Core 2](managed-kafka.md#configuring-seldon-core-2)

## Securing managed Kafka services

You can secure Seldon Core 2 integrations with managed Kafka services using:

* [Kafka Encryption](managed-kafka.md#kafka-encryption-tls)
* [Kafka Authentication](managed-kafka.md#kafka-authentication)

### Kafka Encryption (TLS)

In production settings, always set up TLS encryption with Kafka. This ensures that neither the credentials nor the payloads are transported in plaintext.

{% hint style="info" %}
**Note**: TLS encryption involves only single-sided TLS. This means that the contents are encrypted and sent to the server, but the client won’t send any form of certificate. Therefore, it does not take care of authenticating the client. Client authentication can be configured through mutual TLS (mTLS) or SASL mechanism, which are covered in the [Kafka Authentication](https://docs.seldon.ai/seldon-enterprise-platform/production-environment/kafka/managed-kafka#kafka-authentication) section .
{% endhint %}

When TLS is enabled, the client needs to know the root CA certificate used to create the server’s certificate. This is used to validate the certificate sent back by the Kafka server.

1. Create a certificate named `ca.crt` that is encoded as a PEM certificate. It is important that the certificate is saved as `ca.crt`. Otherwise, Seldon Core 2 may not be able to find the certificate. Within the cluster, you can provide the server’s root CA certificate through a secret. For example, a secret named `kafka-broker-tls` with a certificate.

```
kubectl create secret generic kafka-broker-tls -n seldon --from-file ./ca.crt
```

### Kafka Authentication

In production environments, Kafka clusters often require authentication, especially when using managed Kafka solutions. Therefore, when installing Seldon Core 2 components, it is crucial to provide the correct credentials for a secure connection to Kafka.

The type of authentication used with Kafka varies depending on the setup but typically includes one of the following:

* Simple Authentication and Security Layer (SASL): Requires a username and password.
* Mutual TLS (mTLS): Involves using SSL certificates as credentials.
* OAuth 2.0: Uses the client credential flow to acquire a JWT token.

These credentials are stored as Kubernetes secrets within the cluster. When setting up Seldon Core 2 you must create the appropriate secret in the correct format and update the `components-values.yaml`, and `install-values` files respectively.

{% tabs %}
{% tab title="SASL" %}
When you use SASL as the authentication mechanism for Kafka, the credentials consist of a `username` and `password` pair. The password is supplied through a secret.

{% hint style="info" %}
**Note**:

* Ensure that the field used for the password within the secret is named `password`. Otherwise, Seldon Core 2 may not be able to find the correct password.
* This `password` must be present in the `seldon-logs` namespace (or whichever namespace you wish to use for logging) and every namespace containing Seldon Core 2 runtime.
{% endhint %}

To create a password for Seldon Core 2 in the namespace `seldon`, run the following command:

    ```
    kubectl create secret generic kafka-sasl-secret --from-literal password=<kafka-password> -n seldon
    ```

**Values in Seldon Core 2**

In Seldon Core 2 you need to specify these values in `components-values.yaml`

* `security.kafka.sasl.mechanism` - SASL security mechanism, e.g. `SCRAM-SHA-512`
* `security.kafka.sasl.client.username` - Kafka username
* `security.kafka.sasl.client.secret` - Created secret with `password`
* `security.kafka.ssl.client.brokerValidationSecret` - Certificate Authority of Kafka Brokers

The resulting set of values to include in `components-values.yaml` is similar to:

```
security:
  kafka:
    protocol: SASL_SSL
    sasl:
      mechanism: SCRAM-SHA-512
      client:
        username: <kafka-username>      # TODO: Replace with your Kafka username
        secret: kafka-sasl-secret       # NOTE: Secret name from previous step
    ssl:
      client:
        secret:                                   # NOTE: Leave empty
        brokerValidationSecret: kafka-broker-tls  # NOTE: Optional
```

The `security.kafka.ssl.client.brokerValidationSecret` field is optional. Leave it empty if your brokers use well known Certificate Authority such as Let’s Encrypt.
{% endtab %}

{% tab title="OAuth2.0" %}
When you use OAuth 2.0 as the authentication mechanism for Kafka, the credentials consist of a Client ID and Client Secret, which are used with your Identity Provider to obtain JWT tokens for authenticating with Kafka brokers.

1.  Create a Kubernetes secret `kafka-oauth.yaml` file.

    ```
    apiVersion: v1
    kind: Secret
    metadata:
      name: kafka-oauth
    type: Opaque
    stringData:
      method: OIDC
      client_id: <client id>
      client_secret: <client secret>
      token_endpoint_url: <token endpoint url>
      extensions: ""
      scope: ""
    ```
2.  Store the secret in the `seldon` namespace to configure with Seldon Core 2.

    ```
    kubectl apply -f kafka-oauth.yaml -n seldon
    ```

    This secret must be present in `seldon-logs` namespace and every namespace containing Seldon Core 2 runtime.

Client ID, client secret and token endpoint url should come from identity provider such as Keycloak or Azure AD.

**Values in Seldon Core 2**

In Seldon Core 2 you need to specify these values:

* `security.kafka.sasl.mechanism` - set to `OAUTHBEARER`
* `security.kafka.sasl.client.secret` - Created secret with client credentials
* `security.kafka.ssl.client.brokerValidationSecret` - Certificate Authority of Kafka brokers

The resulting set of values in `components-values.yaml` is similar to:

```
security:
  kafka:
    protocol: SASL_SSL
    sasl:
      mechanism: OAUTHBEARER
      client:
          secret: kafka-oauth                     # NOTE: Secret name from earlier step
    ssl:
      client:
        secret:                                   # NOTE: Leave empty
        brokerValidationSecret: kafka-broker-tls  # NOTE: Optional
```

The `security.kafka.ssl.client.brokerValidationSecret` field is optional. Leave it empty if your brokers use well known Certificate Authority such as Let’s Encrypt.
{% endtab %}

{% tab title="mTLS" %}
When you use `mTLS` as authentication mechanism Kafka uses a set of certificates to authenticate the client.

* A **client certificate**, referred to as `tls.crt`.
* A **client key**, referred to as `tls.key`.
* A **root certificate**, referred to as `ca.crt`.

These certificates are expected to be encoded as PEM certificates and are provided through a secret, which can be created in teh namespace `seldon`:

```
kubectl create secret generic kafka-client-tls -n seldon \
  --from-file ./tls.crt \
  --from-file ./tls.key \
  --from-file ./ca.crt
```

This secret must be present in `seldon-logs` namespace and every namespace containing Seldon Core 2 runtime.

Ensure that the field used within the secret follow the same naming convention: `tls.crt`, `tls.key` and `ca.crt`. Otherwise, Seldon Core 2 may not be able to find the correct set of certificates.

Reference these certificates within the corresponding Helm values for both Seldon Core 2.

**Values for Seldon Core 2** In Seldon Core 2 you need to specify these values:

* `security.kafka.ssl.client.secret` - Secret name containing client certificates
* `security.kafka.ssl.client.brokerValidationSecret` - Certificate Authority of Kafka Brokers

The resulting set of values to include in `components-values.yaml` is similar to:

```
security:
  kafka:
    protocol: SSL
    ssl:
      client:
        secret: kafka-client-tls                  # NOTE: Secret name from earlier step
        brokerValidationSecret: kafka-broker-tls  # NOTE: Optional
```

The `security.kafka.ssl.client.brokerValidationSecret` field is optional. Leave it empty if your brokers use well known Certificate Authority such as Let’s Encrypt.

{% endtab %}
{% endtabs %}

### Example configurations for managed Kafka services

Here are some examples to create secrets for managed Kafka services such as Azure Event Hub, Confluent Cloud(SASL), Confluent Cloud(OAuth2.0).

{% tabs %}
{% tab title="Azure Event Hub with SASL" %}
**Prerequisites**:

* You must use at least the Standard tier for your Event Hub namespace because the Basic tier does not support the Kafka protocol.
* Seldon Core 2 creates two Kafka topics for each model and pipeline, plus one global topic for errors. This results in a total number of topics calculated as: 2 x (number of models + number of pipelines) + 1. This topic count is likely to exceed the limit of the Standard tier in Azure Event Hub. For more information, see [quota information](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers).

**Creating a namespace and obtaining the connection string**

These are the steps that you need to perform in Azure Portal.

1. Create an Azure Event Hub namespace. You need to have an Azure Event Hub namespace. Follow the [Azure quickstart documentation](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-create#create-an-event-hubs-namespace) to create one. **Note**: You do not need to create individual Event Hubs (topics) as Seldon Core 2 automatically creates all necessary topics.
2. Connection string for Kafka Integration. To connect to the Azure Event Hub using the Kafka API, you need to obtain Kafka endpoint and Connection string. For more information, see [Get an Event Hubs connection string](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-get-connection-string?utm\_source=pocket\_saves#connection-string-for-a-namespace)

    **Note**: Ensure you get the Connection string at the namespace level, as it is needed to dynamically create new topics. The format of the Connection string should be:

    ```
    Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=XXXXXX;SharedAccessKey=XXXXXX
    ```

**Creating secrets for Seldon Core 2** 
To store the SASL password in the Kubernetes cluster that run Seldon Core 2, create a secret named `azure-kafka-secret` for Core 2 in the namespace `seldon`. In the following command make sure to replace `<password>` with a password of your choice and `<namespace>` with the namespace form Azure Event Hub.

```
kubectl create secret generic azure-kafka-secret --from-literal=<password>="Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=XXXXXX;SharedAccessKey=XXXXXX" -n seldon
```

{% endtab %}

{% tab title="Confluent Cloud with SASL" %}
**Creating API Keys**

These are the steps that you need to perform in Confluent Cloud.

1. Navigate to **Clients** > **New client** and choose a client, for example **GO** and generate new Kafka cluster API key. For more information, see [Confluent documentation](https://docs.confluent.io/cloud/current/client-apps/config-client.html#configure-clients-from-the-ccloud-console).

Confluent generates a configuration file with the details.

2. Save the values of `Key`, `Secret`, and `bootstrap.servers` from the configuration file.

**Creating secrets for Seldon Core 2**
These are the steps that you need to perform in the Kubernetes cluster that run Seldon Core 2 to store the SASL password.

1. Create a secret named `confluent-kafka-sasl` for Seldon Core 2 in the namespace `seldon`. In the following command make sure to replace `<password>` with with the value of `Secret` that you generated in Confluent cloud.

```
kubectl create secret generic confluent-kafka-sasl --from-literal password="<password>" -n seldon
```
{% endtab %}

{% tab title="Confluent Cloud with OAuth2.0" %}
Confluent Cloud managed Kafka supports OAuth 2.0 to authenticate your Kafka clients. See Confluent Cloud documentation for further details.

**Configuring Identity Provider**

In Confluent Cloud Console Navigate to Account & Access / Identity providers and complete these steps:

1. register your Identity Provider. See Confluent Cloud documentation for further details.
2. add new identity pool to your newly registered Identity Provider. See Confluent Cloud documentation for further details.
3. Obtain these details from Confluent Cloud:
   * Cluster ID: Cluster Overview → Cluster Settings → General → Identification
   * Identity Pool ID: Accounts & access → Identity providers → .
4. Obtain these details from your identity providers such as Keycloak or Azure AD.
   * Client ID
   * Client secret
   * Token Endpoint URL

If you are using Azure AD you may will need to set `scope: api://<client id>/.default`.

**Creating Kubernetes secret**

1.  Create Kubernetes secrets to store the required client credentials. For example, create a `kafka-secret.yaml` file by replacing the values of `<client id>`, `<client secret>`, `<token endpoint url>`, `<cluster id>`,`<identity pool id>` with the values that you obtained from Confluent Cloud and your identity provider.

    ```
    apiVersion: v1
    kind: Secret
    metadata:
      name: confluent-kafka-oauth
    type: Opaque
    stringData:
      method: OIDC
      client_id: <client id>
      client_secret: <client secret>
      token_endpoint_url: <token endpoint url>
      extensions: logicalCluster=<cluster id>,identityPoolId=<identity pool id>
      scope: ""
    ```
2.  Provide the secret named `confluent-kafka-oauth` in the `seldon` namespace to configure with Seldon Core 2.

    ```
    kubectl apply -f kafka-secret.yaml -n seldon
    ```

    This secret must be present in `seldon-logs` namespace and every namespace containing Seldon Core 2 runtime.
{% endtab %}
{% endtabs %}

## Configuring Seldon Core 2

To integrate Kafka with Seldon Core 2.

1. Update the initial configuration.

{% tabs %}
{% tab title="Azure Event Hub with SASL" %}
{% hint style="info" %}
**Note**: In these configurations you may need:

* to tweak the values for `replicationFactor` and `numPartitions` that best suits your cluster configuration.
* set the value for `username` as `$ConnectionString` this is not a variable.
* replace `<namespace>` with the namespace in Azure Event Hub.
{% endhint %}

Update the initial configuration for Seldon Core 2 in the `components-values.yaml` file. Use your preferred text editor to update and save the file with the following content:

```yaml
controller:
  clusterwide: true

dataflow:
  resources:
    cpu: 500m

envoy:
  service:
    type: ClusterIP

kafka:
  bootstrap: <namespace>.servicebus.windows.net:9093
  topics:
    replicationFactor: 3
    numPartitions: 4    
security:
  kafka:
    protocol: SASL_SSL
    sasl:
      mechanism: "PLAIN"
      client:
        username: $ConnectionString
        secret: azure-kafka-secret
    ssl:
      client:
        secret:
        brokerValidationSecret:
        
opentelemetry:
  enable: false

scheduler:
  service:
    type: ClusterIP

serverConfig:
  mlserver:
    resources:
      cpu: 1
      memory: 2Gi

  triton:
    resources:
      cpu: 1
      memory: 2Gi

serviceGRPCPrefix: "http2-"
```
{% endtab %}

{% tab title="Confluent Cloud with SASL" %}
{% hint style="info" %}
**Note**: In these configurations you may need:

* to tweak the values for `replicationFactor` and `numPartitions` that best suits your cluster configuration.
* `replace <username> with the`value of `Key` that you generated in Confluent Cloud.
* replace `<confluent-endpoints>` with the value of `bootstrap.server` that you generated in Confluent Cloud.
{% endhint %}

Update the initial configuration for Seldon Core 2 Operator in the `components-values.yaml` file. Use your preferred text editor to update and save the file with the following content:

```yaml
controller:
  clusterwide: true

dataflow:
  resources:
    cpu: 500m

envoy:
  service:
    type: ClusterIP

kafka:
  bootstrap: <confluent-endpoints>
  topics:
    replicationFactor: 3
    numPartitions: 4
  consumer:
    messageMaxBytes: 8388608
  producer:
    messageMaxBytes: 8388608

security:
  kafka:
    protocol: SASL_SSL
    sasl:
      mechanism: "PLAIN"
      client:
        username: <username>
        secret: confluent-kafka-sasl
    ssl:
      client:
        secret:
        brokerValidationSecret:

opentelemetry:
  enable: false

scheduler:
  service:
    type: ClusterIP

serverConfig:
  mlserver:
    resources:
      cpu: 1
      memory: 2Gi

  triton:
    resources:
      cpu: 1
      memory: 2Gi

serviceGRPCPrefix: "http2-"
```
{% endtab %}

{% tab title="Confluent Cloud with OAuth2.0" %}
{% hint style="info" %}
**Note**: In these configurations you may need:

* to tweak the values for `replicationFactor` and `numPartitions` that best suits your cluster configuration.
* replace `<confluent-endpoints>` with the value of `bootstrap.server` that you generated in Confluent Cloud.
{% endhint %}

Update the initial configuration for Seldon Core 2 Operator in the `components-values.yaml` file. Use your preferred text editor to update and save the file with the following content:

```yaml
controller:
  clusterwide: true

dataflow:
  resources:
    cpu: 500m

envoy:
  service:
    type: ClusterIP

kafka:
  bootstrap: <confluent-endpoints>
  topics:
    replicationFactor: 3
    numPartitions: 4
  consumer:
    messageMaxBytes: 8388608
  producer:
    messageMaxBytes: 8388608

security:
  kafka:
    protocol: SASL_SSL
    sasl:
      mechanism: OAUTHBEARER
      client:
        secret: confluent-kafka-oauth
    ssl:
      client:
        secret:
        brokerValidationSecret:

opentelemetry:
  enable: false

scheduler:
  service:
    type: ClusterIP

serverConfig:
  mlserver:
    resources:
      cpu: 1
      memory: 2Gi

  triton:
    resources:
      cpu: 1
      memory: 2Gi

serviceGRPCPrefix: "http2-"
```
{% endtab %}
{% endtabs %}

2. To enable Kafka Encryption (TLS) you need to reference the [secret](managed-kafka.md#kafka-encryption-tls) that you created in the `security.kafka.ssl.client.secret` field of the Helm chart values. The resulting set of values to include in `components-values.yaml` is similar to:

```
security:
  kafka:
    ssl:
      secret:
        brokerValidationSecret: kafka-broker-tls
```

3. Change to the directory that contains the `components-values.yaml` file and then install Seldon Core 2 operator in the namespace `seldon-system`.

```
 helm upgrade seldon-core-v2-components seldon-charts/seldon-core-v2-setup \
 --version 2.8.5 \
 -f components-values.yaml \
 --namespace seldon-system \
 --install
```
