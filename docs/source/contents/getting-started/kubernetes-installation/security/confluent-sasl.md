# Confluent Cloud SASL Example

> New in Seldon Core 2.5.0

Seldon Core v2 can integrate with Confluent Cloud managed Kafka.
In this example we use SASL security mechanism.


## Create API Keys

In your Confluent Cloud environment create new API keys.
The easiest way to obtain all required information is to head to `Clients` -> `New client` (choose e.g. Go) and generate new Kafka cluster API key from there.

This will generate for you:
- `Key` (we use it as `username`)
- `Secret` (we use it as `password`)

Do not forget to also copy the `bootstrap.servers` from the example config.

See Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/client-apps/config-client.html) in case of issues.

## Create Kubernetes Secret

Seldon Core v2 expects password to be in form of K8s secret
```bash
kubectl create secret generic confluent-kafka-sasl -n seldon-mesh --from-literal password="<Confluent Cloud API Secret>"
```

## Configure Seldon Core v2

Configure Seldon Core v2 by setting following Helm values:

```{literalinclude} ../../../../../../k8s/samples/values-confluent-kafka-sasl.yaml.tmpl
:language: yaml
```

Note you may need to tweak `replicationFactor` and `numPartitions` to your cluster configuration.


## Troubleshooting

- First check Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/overview.html).

- Set the kafka config map debug setting to `all`. For Helm install you can set `kafka.debug=all`.
