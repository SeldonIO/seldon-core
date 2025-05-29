---
description: Learn how to integrate Seldon Core 2 with Confluent Cloud using OAuth 2.0 authentication for secure Kafka communication.
---

# Confluent OAuth Integration

> New in Seldon Core 2.7.0

Seldon Core 2 can integrate with Confluent Cloud managed Kafka.
In this example we use [Oauth 2.0 security mechanism](https://docs.confluent.io/cloud/current/access-management/authenticate/oauth/overview.html).


## Configure Identity Provider in Confluent Cloud Console

In your Confluent Cloud Console go to [Account & Access / Identity providers](https://confluent.cloud/settings/org/identity_providers) and register your Identity Provider.

See Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/access-management/authenticate/oauth/identity-providers.html) for further details.


## Configure Identity Pool

In your Confluent Cloud Console go to [Account & Access / Identity providers](https://confluent.cloud/settings/org/identity_providers) and add new identity pool to your newly registered Identity Provider.

See Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/access-management/authenticate/oauth/identity-pools.html) for further details.

## Create Kubernetes Secret

Seldon Core 2 expects oauth credentials to be in form of K8s secret
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: confluent-kafka-oauth
  namespace: seldon-mesh
type: Opaque
stringData:
  method: OIDC
  client_id: <client id>
  client_secret: <client secret>
  token_endpoint_url: <token endpoint url>
  extensions: logicalCluster=<cluster id>,identityPoolId=<identity pool id>
  scope: ""
```

You need the following information from Confluent Cloud:
- Cluster ID: `Cluster Overview` → `Cluster Settings` → `General` → `Identification`
- Identity Pool ID: `Accounts & access` → `Identity providers` → `<specific provider details>`

Client ID, client secret and token endpoint url should come from identity provider, e.g. Keycloak or Azure AD.


## Configure Seldon Core 2

Configure Seldon Core 2 by setting following Helm values:

```yaml
# k8s/samples/values-confluent-kafka-oauth.yaml.tmpl
kafka:
  bootstrap: < Confluent Cloud Broker Endpoints >
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
```

Note you may need to tweak `replicationFactor` and `numPartitions` to your cluster configuration.


## Troubleshooting

- First check Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/overview.html).

- Set the kafka config map debug setting to `all`. For Helm install you can set `kafka.debug=all`.
