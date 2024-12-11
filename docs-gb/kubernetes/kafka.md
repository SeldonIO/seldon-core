# Kafka

Seldon Core 2 requires [Kafka](https://kafka.apache.org/) to implement data-centric inference Pipelines.
See our [architecture](../architecture/README.md) documentation to learn more on how Seldon Core 2 uses Kafka.


{% hint style="info" %}
**Note**: Kafka integration is required to enable data-centric inference pipelines feature.
It is highly advice to configure Kafka integration to take full advantage of Seldon Core 2 features.
{% endhint %}

We list alternatives below.

## Managed Kafka

We recommend to use managed Kafka solution for production installation.
This allow to take away all the complexity on running secure and scalable Kafka cluster away.

We currently have tested and documented integration with following managed solutions:
- Confluent Cloud (security: SASL/PLAIN)
- Confluent Cloud (security: SASL/OAUTHBEARER)
- Amazon MSK (security: mTLS)
- Amazon MSK (security: SASL/SCRAM)
- Azure Event Hub (security: SASL/PLAIN)

See our [Kafka security](../getting-started/kubernetes-installation/security.md#kafka)
section for configuration examples.

## Self Hosted Kafka

### Strimzi Kafka

Seldon Core 2 requires Kafka to implement data-centric inference Pipelines.
To install Kafka for testing purposed in your k8s cluster, we recommend to use [Strimzi Operator](https://github.com/strimzi/strimzi-kafka-operator).

{% hint style="info" %}
**Note**: This page discuss how to install Strimzi Operator and create Kafka cluster for trial, dev, or testing purposes.
For production grade installation consult [Strimzi documentation](https://strimzi.io/documentation/) or use one of managed solutions mentioned [here](./index.md).
{% endhint %}

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

{% hint style="warning" %}
**Warning**:
Currently Kraft installation of Strimzi is not production ready.
See Strimzi [documentation](https://strimzi.io/docs/operators/0.35.0/deploying.html#ref-operator-use-kraft-feature-gate-str)
and related GitHub [issue](https://github.com/strimzi/strimzi-kafka-operator/issues/5615) for further details.
{% endhint %}


Create Kafka cluster in `seldon-mesh` namespace

```bash
helm upgrade seldon-core-v2-kafka kafka/strimzi -n seldon-mesh --install
```

Note that a specific strimzi operator version is associated with a subset of supported Kafka versions.


## Ansible

We provide automation around the installation of a Kafka cluster for Seldon Core 2 to help with
development and testing use cases.
You can follow the steps defined [here](../getting-started/kubernetes-installation/ansible.md) to
install Kafka via ansible.

You can use our Ansible playbooks to install **only** Strimzi Operator and Kafka cluster by
setting extra Ansible vars:
```bash
ansible-playbook playbooks/setup-ecosystem.yaml -e full_install=no -e install_kafka=yes
```


## Notes
- You can check [kafka-examples](https://github.com/strimzi/strimzi-kafka-operator/tree/main/examples/kafka) for more details.
- As we are using [KRaft](https://kafka.apache.org/documentation/#kraft), use Kafka version 3.4 or above.
- For security settings check [here](../getting-started/kubernetes-installation/security.md#kafka).
