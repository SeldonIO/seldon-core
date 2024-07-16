# Azure Event Hub SASL Example

> New in Seldon Core 2.5.0

Seldon Core v2 can integrate with Azure Event Hub via Kafka protocol.

```{warning}
You will need at least `Standard` tier for your Event Hub Namespace as `Basic` tier does not support Kafka protocol.
```

```{warning}
Seldon Core v2 creates 2 Kafka topics for each pipeline and model plus one global topic for errors.
This means that total number of topics will be `2 x (#models + #pipelines) + 1` which will likely exceed the limit of `Standard` tier in Azure Event Hub.
You can find more information on quotas, like the number of partitions per Event Hub, [here](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-quotas#basic-vs-standard-vs-premium-vs-dedicated-tiers).
```

## Prerequisites

To start you will need to have an Azure Event Hub Namespace.
You can create one following Azure quickstart [docs](https://learn.microsoft.com/en-gb/azure/event-hubs/event-hubs-create).
Note that you do not need to create an Event Hub (topics) as Core v2 will require all the topics it needs automatically.

## Create API Keys

To connect to Azure Event Hub provided Kafka API you need to obtain:
- Kafka Endpoint
- Connection String

You can obtain both using Azure Portal as documented [here](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-get-connection-string?utm_source=pocket_saves#connection-string-for-a-namespace).

```{note}
You should get the Connection String for a namespace level as we will need to dynamically create new topics.
```

The Connection String should be in format of
```
Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=XXXXXX;SharedAccessKey=XXXXXX
```

## Create Kubernetes Secret

Seldon Core v2 expects password to be in form of K8s secret
```bash
kubectl create secret generic azure-kafka-secret -n seldon-mesh --from-literal password="Endpoint=sb://<namespace>.servicebus.windows.net/;SharedAccessKeyName=XXXXXX;SharedAccessKey=XXXXXX"
```

## Configure Seldon Core v2

Configure Seldon Core v2 by setting following Helm values:

```{literalinclude} ../../../../../../k8s/samples/values-azure-event-hub-sasl.yaml.tmpl
:language: yaml
```

Note you may need to tweak `replicationFactor` and `numPartitions` to your cluster configuration. The username should read `$ConnectionString` and this is not a variable for you to replace.

## Troubleshooting

- First check Azure Event Hub [troubleshooting guide](https://learn.microsoft.com/en-us/azure/event-hubs/troubleshooting-guide).

- Set the kafka config map debug setting to `all`. For Helm install you can set `kafka.debug=all`.

- Verify that you did not hit quotas for topics or partitions in your Event Hub namespace.
