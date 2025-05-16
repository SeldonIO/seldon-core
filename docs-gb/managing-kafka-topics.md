# Managing Kafka Topics in Seldon Core 2

## Model Kafka topics

A [Model](./kubernetes/resources/model.md) in Seldon Core 2 represents the fundamental unit for serving a machine learning artifact within a running server instance.

If Kafka is installed in your cluster, Seldon Core automatically creates dedicated input and output topics for each model as it is loaded. These topics facilitate asynchronous messaging, enabling clients to send input messages and retrieve output responses independently and at a later time.

By default, when a model is unloaded, the associated Kafka topics are preserved. This supports use cases like auditing, but can also lead to increased Kafka resource usage and unnecessary costs for workloads that don't require persistent topics.

You can control this behavior by configuring the `dataflow` section of the model specification. Alongside the required `storageUri`, and `requirements` fields, you can optionally include the `cleanTopicsOnDelete` flag. This boolean setting determines whether the associated Kafka topics should be deleted when the model is unloaded:

* When set to `false` (the default), the topics remain after the model is deleted.
* When set to `true`, both the input and output topics are removed when the model is unloaded.

Here is an example of a manifest file that enables topic cleanup on deletion:
```yaml
# samples/models/sklearn-iris-gs.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  dataflow:
    cleanTopicsOnDelete: true
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.5.0/iris-sklearn"
  requirements:
    - sklearn
  memory: 100Ki
```

To inspect existing Kafka topics in your cluster, you can deploy a temporary Pod:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: kafka-busybox
spec:
  containers:
    - name: kafka-busybox
      image: apache/kafka:latest
      command: ["sleep", "3600"]
      imagePullPolicy: IfNotPresent
  restartPolicy: Always
```

After the Pod is running, you can access it and list topics with the following command:

```bash
kafka-busybox:/opt/kafka/bin$ ./kafka-topics.sh --list --bootstrap-server $SELDON_KAFKA_BOOTSTRAP_PORT_9092_TCP
```

## Deploying and verifying topic cleanup for a model

Apply the model manifest with topic cleanup enabled:
```bash
kubectl apply -f model.yaml -n seldon-mesh
```

After deployment, you can list Kafka topics from within the `kafka-busybox` pod and confirm that input/output topics have been created:
```bash
__consumer_offsets
seldon.seldon-mesh.model.iris.inputs
seldon.seldon-mesh.model.iris.outputs
```

To delete the model:
```bash
kubectl delete -f model.yaml -n seldon-mesh
```

After deletion, list the topics again. You should see that the input and output topics have been successfully removed from Kafka:
```bash
__consumer_offsets
```

## Pipeline Kafka topics

Similar to models, when a [Pipeline](./kubernetes/resources/pipeline.md) is deployed in Seldon Core 2, Kafka input and output topics are automatically created for it. These topics enable asynchronous processing across pipeline steps.

As with models, the `cleanTopicsOnDelete` flag controls whether these topics are retained or removed when the pipeline is deleted:

* By default, topics are retained after the pipeline is unloaded.

* When `cleanTopicsOnDelete` is set to `true`, the input and output topics associated with the pipeline are deleted.

Here is an example of a pipeline manifest that wraps the previously defined model and enables topic cleanup:
```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: iris-pipeline
spec:
  dataflow:
    cleanTopicsOnDelete: true
  steps:
    - name: iris
  output:
    steps:
    - iris
```

## Deploying and verifying topic cleanup for a pipeline

Apply the pipeline manifest with topic cleanup enabled:
```bash
kubectl apply -f pipeline.yaml -n seldon-mesh
```

After the pipeline is deployed, you can list the Kafka topics from inside the `kafka-busybox` pod to confirm that they have been created:
```bash
__consumer_offsets
seldon.seldon-mesh.errors.errors
seldon.seldon-mesh.model.iris.inputs
seldon.seldon-mesh.model.iris.outputs
seldon.seldon-mesh.pipeline.iris-pipeline.inputs
seldon.seldon-mesh.pipeline.iris-pipeline.outputs
```

To delete the pipeline, run:
```bash
kubectl delete -f pipeline.yaml -n seldon-mesh
```

After deletion, list the Kafka topics again. You should observe that the pipeline’s input and output topics have been removed:
```bash
__consumer_offsets
seldon.seldon-mesh.errors.errors
seldon.seldon-mesh.model.iris.inputs
seldon.seldon-mesh.model.iris.outputs
```

{% hint style="info" %}
Note: Topics associated with models used inside the pipeline are not deleted unless the corresponding models are also unloaded and the `cleanTopicsOnDelete` flag was set to `true` in their specification.
{% endhint %}