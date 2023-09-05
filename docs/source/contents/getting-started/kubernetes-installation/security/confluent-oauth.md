# Confluent Cloud Oauth 2.0 Example

> New in Seldon Core 2.7.0

Seldon Core v2 can integrate with Confluent Cloud managed Kafka.
In this example we use [Oauth 2.0 security mechanism](https://docs.confluent.io/cloud/current/access-management/authenticate/oauth/overview.html).


## Configure Identity Provider in Confluent Cloud Console

In your Confluent Cloud Console go to [Account & Access / Identity providers](https://confluent.cloud/settings/org/identity_providers) and register your Identity Provider.

See Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/access-management/authenticate/oauth/identity-providers.html) for further details.


## Configure Identity Pool

In your Confluent Cloud Console go to [Account & Access / Identity providers](https://confluent.cloud/settings/org/identity_providers) and add new identity pool to your newly registered Identity Provider.

See Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/access-management/authenticate/oauth/identity-pools.html) for further details.


## Create Kubernetes Secret

Seldon Core v2 expects oauth credentials to be in form of K8s secret
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

You will need following information from Confluent Cloud:
- Cluster ID: `Cluster Overview` → `Cluster Settings` → `General` → `Identification`
- Identity Pool ID: `Accounts & access` → `Identity providers` → `<specific provider details>`

Client ID, client secret and token endpoint url should come from identity provider, e.g. Keycloak or Azure AD.


## Configure Seldon Core v2

Configure Seldon Core v2 by setting following Helm values:

```{literalinclude} ../../../../../../k8s/samples/values-confluent-kafka-oauth.yaml.tmpl
:language: yaml
```

Note you may need to tweak `replicationFactor` and `numPartitions` to your cluster configuration.


## Troubleshooting

- First check Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/overview.html).

- Set the kafka config map debug setting to `all`. For Helm install you can set `kafka.debug=all`.
