---
description: Integrate self-hosted Kafka with Seldon Core 2.
---
# Self-hosted Kafka

You can run Kafka in the same Kubernetes cluster that hosts the Seldon Core 2. Seldon recommends using the [Strimzi operator](https://strimzi.io/docs/operators/latest/deploying) for Kafka installation and maintenance. For more details about configuring Kafka with Seldon Core 2 see the [Configuration](../../getting-started/configuration.md)section.

{% hint style="info" %}
**Note**: These instructions help you quickly set up a Kafka cluster. For production grade installation consult [Strimzi documentation](https://strimzi.io/documentation/) or use one of [managed solutions](../production-environment/#managed-kafka). 
{% endhint %}

Integrating self-hosted Kafka with Seldon Core 2 includes these steps:

1. [Install Kafka](self-hosted-kafka.md#installing-kafka-in-a-kubernetes-cluster)
2. [Configure Seldon Core 2](self-hosted-kafka.md#configuring-seldon-core-2)

## Installing Kafka in a Kubernetes cluster

Strimzi provides a Kubernetes Operator to deploy and manage Kafka clusters. First, we need to install the Strimzi Operator in your Kubernetes cluster.

1.  Create a namespace where you want to install Kafka. For example the namespace `seldon-mesh`:

    ```
    kubectl create namespace seldon-mesh || echo "namespace seldon-mesh exists"
    ```
2.  Install Strimzi.

    ```
    helm repo add strimzi https://strimzi.io/charts/
    helm repo update
    ```
3.  Install Strimzi Operator.

    ```
    helm install strimzi-kafka-operator strimzi/strimzi-kafka-operator --namespace seldon-mesh
    ```

    This deploys the `Strimzi Operator` in the `seldon-mesh` namespace. After the Strimzi Operator is running, you can create a Kafka cluster by applying a Kafka custom resource definition.
4.  Create a YAML file to specify the initial configuration. For example, create the `kafka.yaml` file. Use your preferred text editor to create and save the file with the following content:

    ```yaml
     apiVersion: kafka.strimzi.io/v1beta2
     kind: Kafka
     metadata:
       name: seldon
       namespace: seldon-mesh
     spec:
       kafka:
         replicas: 3
         version: 3.7.0
         config:
           auto.create.topics.enable: true
           default.replication.factor: 1
           inter.broker.protocol.version: 3.7
           min.insync.replicas: 1
           offsets.topic.replication.factor: 1
           transaction.state.log.min.isr: 1
           transaction.state.log.replication.factor: 1
         listeners:
         - name: plain
           port: 9092
           tls: false
           type: internal
         storage:
           type: ephemeral
       zookeeper:
         replicas: 1
         storage:
           type: ephemeral
    ```

    This configuration sets up a Kafka cluster with version 3.7.0. Ensure that you review the the [supported versions](https://strimzi.io/downloads/) of Kafka and update the version in the `kafka.yaml` file as needed. For more configuration examples, see this [strimzi-kafka-operator](https://github.com/strimzi/strimzi-kafka-operator/tree/main/examples/kafka).

5.  Apply the Kafka cluster configuration.

    ```
    kubectl apply -f kafka.yaml -n seldon-mesh
    ```
6.  Check the status of the Kafka Pods to ensure they are running properly:

    ```
    kubectl get pods -n seldon-mesh
    ```

    You should see multiple Pods for Kafka, Zookeeper, and Strimzi operators running.

    ```
    NAME                                            READY   STATUS    RESTARTS        AGE
    hodometer-749d7c6875-4d4vw                      1/1     Running   0               17m
    mlserver-0                                      3/3     Running   0               16m
    seldon-dataflow-engine-7b98c76d67-v2ztq         1/1     Running   8 (5m33s ago)   17m
    seldon-envoy-bb99f6c6b-4mpjd                    1/1     Running   0               17m
    seldon-kafka-0                                  1/1     Running   0               111s
    seldon-kafka-1                                  1/1     Running   0               111s
    seldon-kafka-2                                  1/1     Running   0               111s
    seldon-modelgateway-5c76c7695b-bhfj5            1/1     Running   0               17m
    seldon-pipelinegateway-584c7d95c-bs8c9          1/1     Running   0               17m
    seldon-scheduler-0                              1/1     Running   0               17m
    seldon-v2-controller-manager-5dd676c7b7-xq5sm   1/1     Running   0               17m
    seldon-zookeeper-0                              1/1     Running   0               2m26s
    strimzi-cluster-operator-7cf9ff5686-6tb7p       1/1     Running   0               5m10s
    triton-0                                        3/3     Running   0               16m
    ```

## Configuring Seldon Core 2

When the `SeldonRuntime` is installed in a namespace a ConfigMap is created with the
settings for Kafka configuration.
{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/scheduler/config/kafka-internal.json" %} 

1. Verify that the ConfigMap resource named `seldon-kafka` that is created in the namespace `seldon-mesh`:

    ```
    kubectl get configmaps -n seldon-mesh
    ```

    You should the ConfigMaps for Kafka, Zookeeper, Strimzi operators, and others.

    ```
    NAME                       DATA   AGE
    kube-root-ca.crt           1      50m
    seldon-agent               1      48m
    seldon-kafka               1      48m
    seldon-manager-config      1      49m
    seldon-tracing             4      48m
    seldon-zookeeper-config    2      5m
    strimzi-cluster-operator   1      44m
    ```
2. View the configuration of the the ConfigMap named `seldon-kafka`.

    ```
    kubectl get configmap seldon-kafka -n seldon-mesh -o yaml
    ```

    You should see an output simialr to this:

    ```
    apiVersion: v1
    data:
      kafka.json: '{"bootstrap.servers":"seldon-kafka-bootstrap.seldon-mesh:9092","consumer":{"auto.offset.reset":"earliest","message.max.bytes":"1000000000","session.timeout.ms":"6000","topic.metadata.propagation.max.ms":"300000"},"producer":{"linger.ms":"0","message.max.bytes":"1000000000"},"topicPrefix":"seldon"}'
    kind: ConfigMap
    metadata:
      creationTimestamp: "2024-12-05T07:12:57Z"
      name: seldon-kafka
      namespace: seldon-mesh
      ownerReferences:
      - apiVersion: mlops.seldon.io/v1alpha1
        blockOwnerDeletion: true
        controller: true
        kind: SeldonRuntime
        name: seldon
        uid: 9e724536-2487-487b-9250-8bcd57fc52bb
      resourceVersion: "778"
      uid: 5c041e69-f36b-4f14-8f0d-c8790003cb3e
    ``` 

### Customzing the settings

To customize the settings you can add and modify the Kafka configuration using Helm, for example to add compression for producers.


1. Create a YAML file to specify the compression configuration for Seldon Core 2 runtime. For example, create the `values-runtime-kafka-compression.yaml` file. Use your preferred text editor to create and save the file with the following content:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/k8s/samples/values-runtime-kafka-compression.yaml" %}

2. Change to the directory that contains the `values-runtime-kafka-compression.yaml` file and then install Seldon Core 2 runtime in the namespace `seldon-mesh`.

  ```bash
  helm upgrade seldon-v2-runtime k8s/helm-charts/seldon-core-v2-runtime \
  --namespace seldon-mesh \
  --f values-runtime-kafka-compression.yaml \
  --install
  ```
### Configuring topic and consumer isolation

If you are using a shared Kafka cluster with other applications, it is advisable to isolate topic names and consumer group IDs from other cluster users to prevent naming conflicts. This can be achieved by configuring the following two settings:

* `topicPrefix`: set a prefix for all topics
* `consumerGroupIdPrefix`: set a prefix for all consumer groups

Here’s an example of how to configure topic name and consumer group ID isolation during a Helm installation for an application named `myorg`:

```bash
helm upgrade --install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh \
--set controller.clusterwide=true \
--set kafka.topicPrefix=myorg \
--set kafka.consumerGroupIdPrefix=myorg
```


