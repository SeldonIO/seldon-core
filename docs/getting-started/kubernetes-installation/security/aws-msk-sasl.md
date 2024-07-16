# AWS MSK SASL

> New in Seldon Core 2.5.0

Seldon Core v2 can integrate with Amazon managed Apache Kafka (MSK). You can control access to your Amazon MSK clusters using sign-in credentials that are stored and secured using AWS Secrets Manager. Storing user credentials in Secrets Manager reduces the overhead of cluster authentication such as auditing, updating, and rotating credentials. Secrets Manager also lets you share user credentials across clusters.

```{Note}
Configuration of the AWS MSK instance itself is out of scope for this example.
Please follow the official [AWS documentation](https://docs.aws.amazon.com/msk/latest/developerguide/what-is-msk.html) on how to enable SASL and public access to the Kafka cluster (if required).
```

## Setting up SASL/SCRAM authentication for an Amazon MSK cluster

To setup SASL/SCRAM in an Amazon MSK cluster, please follow the guide from Amazon's Official [documentation](https://docs.aws.amazon.com/msk/latest/developerguide/msk-password.html#msk-password-tutorial).

Do not forget to also copy the `bootstrap.servers` which we will use it in our configuration later below for Seldon.

## Create Kubernetes Secret

Seldon Core v2 expects password to be in form of K8s secret
```bash
kubectl create secret generic aws-msk-kafka-secret -n seldon-mesh --from-literal password="<MSK SASL Password>"
```

## Configure Seldon Core v2

Configure Seldon Core v2 by setting following Helm values:

```{literalinclude} ../../../../../../k8s/samples/values-aws-msk-kafka-sasl-scram.yaml.tmpl
:language: yaml
```

Note you may need to tweak `replicationFactor` and `numPartitions` to your cluster configuration.



## Troubleshooting

- Please check Amazon MSK Troubleshooting [documentation](https://docs.aws.amazon.com/msk/latest/developerguide/troubleshooting.html).

- Set the kafka config map debug setting to `all`. For Helm install you can set `kafka.debug=all`.
