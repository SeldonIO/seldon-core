---
description: Learn how to set up and configure self-hosted Kafka with Seldon Core 2 using Strimzi operator in Kubernetes. This comprehensive guide covers Kafka cluster deployment, node pool configuration, TLS encryption, topic isolation, and Helm customization for production-ready ML model serving.
---
# Self-hosted Kafka

You can run Kafka in the same Kubernetes cluster that hosts the Seldon Core 2. Seldon recommends using the [Strimzi operator](https://strimzi.io/docs/operators/latest/deploying) for Kafka installation and maintenance. For more details about configuring Kafka with Seldon Core 2 see the [Configuration](../../getting-started/configuration.md) section.

{% hint style="info" %}
**Note**: These instructions help you quickly set up a Kafka cluster. For production grade installation consult [Strimzi documentation](https://strimzi.io/documentation/) or use one of [managed solutions](../production-environment/kafka/README.md). 
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
4.  Create a YAML file to specify the initial configuration.
 
    **Note**: This configuration sets up a Kafka cluster with version 3.9.0. Ensure that you review the the [supported versions](https://strimzi.io/downloads/) of Kafka and update the version in the `kafka.yaml` file as needed. For more configuration examples, see this [strimzi-kafka-operator](https://github.com/strimzi/strimzi-kafka-operator/tree/main/examples/kafka).
    
    Use your preferred text editor to create and save the file as `kafka.yaml` with the following content:

  ```yaml
  apiVersion: kafka.strimzi.io/v1beta2
  kind: Kafka
  metadata:
    name: seldon
    namespace: seldon-mesh
    annotations:
      strimzi.io/node-pools: enabled
      strimzi.io/kraft: enabled
  spec:
    kafka:
      replicas: 3
      version: 3.9.0
      listeners:
        - name: plain
          port: 9092
          tls: false
          type: internal
        - name: tls
          port: 9093
          tls: true
          type: internal
      config:
        processMode: kraft
        auto.create.topics.enable: true
        default.replication.factor: 1
        inter.broker.protocol.version: 3.7
        min.insync.replicas: 1
        offsets.topic.replication.factor: 1
        transaction.state.log.min.isr: 1
        transaction.state.log.replication.factor: 1
    entityOperator: null
  ```
6.  Apply the Kafka cluster configuration.

    ```
    kubectl apply -f kafka.yaml -n seldon-mesh
    ```
7.  Create a YAML file named `kafka-nodepool.yaml` to create a nodepool for the kafka cluster.

  ```yaml
  apiVersion: kafka.strimzi.io/v1beta2
  kind: KafkaNodePool
  metadata:
    name: kafka
    namespace: seldon-mesh
    labels:
      strimzi.io/cluster: seldon
  spec:
    replicas: 3
    roles:
      - broker
      - controller
    resources:
      requests:
        cpu: '500m'
        memory: '2Gi'
      limits:
        memory: '2Gi'
    template:
      pod:
        tmpDirSizeLimit: 1Gi
    storage:
      type: jbod
      volumes:
        - id: 0
          type: ephemeral
          sizeLimit: 500Mi
          kraftMetadata: shared
        - id: 1
          type: persistent-claim
          size: 10Gi
          deleteClaim: false
  ```
8.  Apply the Kafka node pool configuration.

    ```
    kubectl apply -f kafka-nodepool.yaml -n seldon-mesh
    ```  
9.  Check the status of the Kafka Pods to ensure they are running properly:

    ```
    kubectl get pods -n seldon-mesh
    ```
    
{% hint style="info" %}
**Note**: It might take a couple of minutes for all the Pods to be ready.
To check the status of the Pods in real time use this command: `kubectl get pods -w -n seldon-mesh`. 
{% endhint %}

You should see multiple Pods for Kafka, and Strimzi operators running.

    ```bash
    NAME                                            READY   STATUS    RESTARTS      AGE
    hodometer-5489f768bf-9xnmd                      1/1     Running   0             25m
    mlserver-0                                      3/3     Running   0             24m
    seldon-dataflow-engine-75f9bf6d8f-2blgt         1/1     Running   5 (23m ago)   25m
    seldon-envoy-7c764cc88-xg24l                    1/1     Running   0             25m
    seldon-kafka-0                                  1/1     Running   0             21m
    seldon-kafka-1                                  1/1     Running   0             21m
    seldon-kafka-2                                  1/1     Running   0             21m
    seldon-modelgateway-54d457794-x4nzq             1/1     Running   0             25m
    seldon-pipelinegateway-6957c5f9dc-6blx6         1/1     Running   0             25m
    seldon-scheduler-0                              1/1     Running   0             25m
    seldon-v2-controller-manager-7b5df98677-4jbpp   1/1     Running   0             25m
    strimzi-cluster-operator-66b5ff8bbb-qnr4l       1/1     Running   0             23m
    triton-0                                        3/3     Running   0             24m
    ```

### Troubleshooting

**Error**
 The Pod that begins with the name `seldon-dataflow-engine` does not show the status as `Running`.

 One of the possible reasons could be that the DNS resolution for the service failed.

**Solution**
1. Check the logs of the Pod `<seldon-dataflow-engine>`:
   ```
   kubectl logs <seldon-dataflow-engine> -n seldon-mesh
   ```
1. In the output check if a message reads:
   ```
   WARN [main] org.apache.kafka.clients.ClientUtils : Couldn't resolve server seldon-kafka-bootstrap.seldon-mesh:9092 from bootstrap.servers as DNS resolution failed for seldon-kafka-bootstrap.seldon-mesh
   ```
1. Verify the `name` in the `metadata` for the `kafka.yaml` and `kafka-nodepool.yaml`. It should read `seldon`.
1. Check the name of the Kafka services in the namespace:
   ```
   kubectl get svc -n seldon-mesh
   ```
1. Restart the Pod:
   ```
   kubectl delete pod <seldon-dataflow-engine> -n seldon-mesh 
   ```
   
## Configuring Seldon Core 2

When the `SeldonRuntime` is installed in a namespace a ConfigMap is created with the
settings for Kafka configuration. Update the `ConfigMap` only if you need to customize the configurations.
{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/scheduler/config/kafka-internal.json" %} 

1. Verify that the ConfigMap resource named `seldon-kafka` that is created in the namespace `seldon-mesh`:

    ```
    kubectl get configmaps -n seldon-mesh
    ```

    You should the ConfigMaps for Kafka, Zookeeper, Strimzi operators, and others.

    ```
    NAME                       DATA   AGE
    kube-root-ca.crt           1      38m
    seldon-agent               1      30m
    seldon-kafka               1      30m
    seldon-kafka-0             6      26m
    seldon-kafka-1             6      26m
    seldon-kafka-2             6      26m
    seldon-manager-config      1      30m
    seldon-tracing             4      30m
    strimzi-cluster-operator   1      28m
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

After you integrated Seldon Core 2 with Kafka, you need to [Install an Ingress Controller](../production-environment/ingress-controller/istio.md) that adds an abstraction layer for traffic routing by receiving traffic from outside the Kubernetes platform and load balancing it to Pods running within the Kubernetes cluster.

### Customizing the settings (optional)

To customize the settings you can add and modify the Kafka configuration using Helm, for example to add compression for producers.


1. Create a YAML file to specify the compression configuration for Seldon Core 2 runtime. For example, create the `values-runtime-kafka-compression.yaml` file. Use your preferred text editor to create and save the file with the following content:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/k8s/samples/values-runtime-kafka-compression.yaml" %}

2. Change to the directory that contains the `values-runtime-kafka-compression.yaml` file and then install Seldon Core 2 runtime in the namespace `seldon-mesh`.

  ```bash
  helm upgrade seldon-core-v2-runtime seldon-charts/seldon-core-v2-runtime \
  --namespace seldon-mesh \
  -f values-runtime-kafka-compression.yaml \
  --install
  ```
### Configuring topic and consumer isolation (optional)

If you are using a shared Kafka cluster with other applications, it is advisable to isolate topic names and consumer group IDs from other cluster users to prevent naming conflicts. This can be achieved by configuring the following two settings:

* `topicPrefix`: set a prefix for all topics
* `consumerGroupIdPrefix`: set a prefix for all consumer groups

Here's an example of how to configure topic name and consumer group ID isolation during a Helm installation for an application named `myorg`:

```bash
helm upgrade --install seldon-core-v2-setup seldon-charts/seldon-core-v2-setup \
--namespace seldon-mesh \
--set controller.clusterwide=true \
--set kafka.topicPrefix=myorg \
--set kafka.consumerGroupIdPrefix=myorg
```
## Next Steps

After you installed Seldon Core 2, and Kafka using Helm, you need to complete [Installing a Service mesh](../production-environment/ingress-controller/istio.md).

