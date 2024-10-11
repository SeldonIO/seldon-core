# Confluent Cloud SASL Example

> New in Seldon Core 2.5.0

Seldon Core 2 can integrate with Confluent Cloud managed Kafka.
In this example we use SASL security mechanism.


## Create API Keys

In your Confluent Cloud environment create new API keys.
The easiest way to obtain all required information is to head to `Clients` -> `New client`
(choose e.g. Go) and generate new Kafka cluster API key from there.

This will generate for you:
- `Key` (we use it as `username`)
- `Secret` (we use it as `password`)

Do not forget to also copy the `bootstrap.servers` from the example config.

See Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/client-apps/config-client.html) in case of issues.

## Create Kubernetes Secret

Seldon Core 2 expects password to be in form of K8s secret
```bash
kubectl create secret generic confluent-kafka-sasl -n seldon-mesh --from-literal password="<Confluent Cloud API Secret>"
```

## Configure Seldon Core 2

Configure Seldon Core 2 by setting following Helm values:

```yaml
# k8s/samples/values-confluent-kafka-sasl.yaml.tmpl
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
      mechanism: "PLAIN"
      client:
        username: < username >
        secret: confluent-kafka-sasl
    ssl:
      client:
        secret:
        brokerValidationSecret:
```

{% hint style="info" %}
You may need to tweak `replicationFactor` and `numPartitions` to your cluster configuration.
{% endhint %}

## Troubleshooting

- First check Confluent Cloud [documentation](https://docs.confluent.io/cloud/current/overview.html).
- Set the kafka config map debug setting to `all`. For Helm install you can set `kafka.debug=all`.
