# Strimzi Kafka

Seldon Core v2 requires Kafka to implement data-centric inference Pipelines.
To install Kafka for testing purposed in your k8s cluster, we recommend to use [Strimzi Operator](https://github.com/strimzi/strimzi-kafka-operator).

```{note}
This page discuss how to install Strimzi Operator and create Kafka cluster for trial, dev, or testing purposes.
For production grade installation consult [Strimzi documentation](https://strimzi.io/documentation/) or use one of managed solutions mentioned [here](./index.md).
```

You can install and configure Strimzi using either Helm charts or our Ansible playbooks, both documented below.

## Helm

The installation of a Kafka cluster requires the Strimzi Kafka operator installed in the same namespace.
This allows to directly use the mTLS certificates created by Strimzi Operator.
One option to install the Strimzi operator is via [Helm](https://strimzi.io/docs/operators/in-development/full/deploying.html#deploying-cluster-operator-helm-chart-str).

Note that we are using here KRaft instead of Zookeeper for Kafka.
You can enable `featureGates` during Helm installation via:

```bash
helm upgrade --install strimzi-kafka-operator \
  strimzi/strimzi-kafka-operator \
  --namespace seldon-mesh --create-namespace \
  --set featureGates='+UseKRaft\,+UseStrimziPodSets'
```

```{warning}
Use with caution!
Currently Kraft installation of Strimzi is not production ready.
See Strimzi [documentation](https://strimzi.io/docs/operators/0.35.0/deploying.html#ref-operator-use-kraft-feature-gate-str) and related GitHub [issue](https://github.com/strimzi/strimzi-kafka-operator/issues/5615) for further details.
```

Create Kafka cluster in `seldon-mesh` namespace

```bash
helm upgrade seldon-core-v2-kafka kafka/strimzi -n seldon-mesh --install
```

Note that a specific strimzi operator version is associated with a subset of supported Kafka versions.


## Ansible

We provide automation around the installation of a Kafka cluster for Seldon Core v2 to help with development and testing use cases.
You can follow the steps defined [here](../../docs/source/contents/getting-started/kubernetes-installation/ansible.md) to install Kafka via ansible.

You can use our Ansible playbooks to install **only** Strimzi Operator and Kafka cluster by setting extra Ansible vars:
```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e full_install=no -e install_kafka=yes
```


## Notes
- You can check [kafka-examples](https://github.com/strimzi/strimzi-kafka-operator/tree/main/examples/kafka) for more details.
- As we are using [KRaft](https://kafka.apache.org/documentation/#kraft), use Kafka version 3.3 or above.
- For security settings check [here](../../docs/source/contents/getting-started/kubernetes-installation/security/index.md#kafka).
