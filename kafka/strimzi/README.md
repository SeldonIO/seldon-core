# Kafka Integration

Seldon Core v2 relies on Kafka to implement datacentric inference Pipleines. 

To install kafka in your k8s cluster, we higlight two options, ansible and helm.


## Install Kafka

### Ansible

We provide automation around the installation of a Kafka cluster for Seldon Core v2 to help with development and testing usecases.

You can follow the steps defined [here](../../ansible/README.md) to install kafka via ansible.

### Helm

Use helm if you require changes and customisations to the Kafka cluster.

The installation of a kafka cluster requires the strimzi operator installed in the same namespace.

One option to install the strimzi operator is via [helm](https://strimzi.io/docs/operators/in-development/full/deploying.html#deploying-cluster-operator-helm-chart-str)

Note that we recommend using KRaft instead of Zookeeper for Kafka. To enable KRaft set `featureGates` during installation via `helm`

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

