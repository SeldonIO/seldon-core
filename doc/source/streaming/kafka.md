# Native Kafka Stream Processing (>=1.2.0, alpha)

Seldon provides a native kafka integration from version 1.2. when you specify `serverType: kafka` in your SeldonDeployment.

When `serverType: kafka` is specified you need to also specify environment variables in `svcOrchSpec` for KAFKA_BROKER, KAFKA_INPUT_TOPIC, KAFKA_OUTPUT_TOPIC. An example is shown below for a Tensorflow CIFAR10  model:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: tfserving-cifar10
spec:
  protocol: tensorflow
  transport: rest
  serverType: kafka  
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - args: 
          - --port=8500
          - --rest_api_port=8501
          - --model_name=resnet32
          - --model_base_path=gs://seldon-models/tfserving/cifar10/resnet32
          - --enable_batching
          image: tensorflow/serving
          name: resnet32
          ports:
          - containerPort: 8501
            name: http
    svcOrchSpec:
      env:
      - name: KAFKA_BROKER
        value: 10.12.10.16:9094
      - name: KAFKA_INPUT_TOPIC
        value: cifar10-rest-input
      - name: KAFKA_OUTPUT_TOPIC
        value: cifar10-rest-output
    graph:
      name: resnet32
      type: MODEL
      endpoint:
        service_port: 8501
    name: model
    replicas: 1
```

The above creates a REST tensorflow deployment using the tensorflow protocol and connects to input and output topics.

## Details

For the SeldonDeployment:

 1. Start with any Seldon inference graph
 1. Set `spec.serverType` to `kafka`
 1. Add a `spec.predictor[].svcOrchSpec.env` with settings for KAFKA_BROKER, KAFKA_INPUT_TOPIC, KAFKA_OUTPUT_TOPIC.

For the input kafka topic:

Create requests streams for the input prediction of your specified protocol and transport.

 * For REST: the JSON representation of a predict request in the given protocol.
 * For gRPC: the protobuffer binary serialization of the request for the given protocol. You should also add a metadata field called `proto-name` with the package name of the protobuffer so it can be decoded, for example `tensorflow.serving.PredictRequest`. We can only support proto buffers for native grpc protocols supported by Seldon.


## Examples

 [A worked example for a CIFAR10 image classifier is available](../examples/cifar10_kafka.html).
