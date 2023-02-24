# Kafka Integration

Seldon Core v2 requires Kafka to implement data-centric inference Pipelines.
To install Kafka for testing purposed in your k8s cluster, we higlight different options:
## Helm

The installation of a Kafka cluster requires the Strimzi Kafka operator installed in the same namespace.

One option to install the Strimzi operator is via [Helm](https://strimzi.io/docs/operators/in-development/full/deploying.html#deploying-cluster-operator-helm-chart-str)

Note that we recommend using KRaft instead of Zookeeper for Kafka. To enable KRaft set `featureGates` during installation via `helm`.

```bash
helm upgrade --install strimzi-kafka-operator  \
  strimzi/strimzi-kafka-operator \
  --namespace seldon-mesh --create-namespace \
  --set featureGates='+UseKRaft\,+UseStrimziPodSets'
```

Create Kafka cluster in `seldon-mesh` namespace

```bash
helm upgrade seldon-core-v2-kafka kafka/strimzi -n seldon-mesh --install
```

Note that a specific strimzi operator version is assciated with a subset of supported Kafka versions.

## Ansible

We provide automation around the installation of a Kafka cluster for Seldon Core v2 to help with development and testing usecases.

You can follow the steps defined [here](../../ansible/README.md) to install kafka via ansible.

# Notes
- You can check [kafka-examples](https://github.com/strimzi/strimzi-kafka-operator/tree/main/examples/kafka) for more details.
- As we recommned using [KRaft](https://kafka.apache.org/documentation/#kraft), use Kafka version 3.3 or above.
- For security settings check [here](../../docs/source/contents/getting-started/kubernetes-installation/security/index.md#kafka).
