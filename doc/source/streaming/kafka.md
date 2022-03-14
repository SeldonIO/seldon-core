# Native Kafka Stream Processing

Seldon provides a native kafka integration when you specify `serverType: kafka` in your SeldonDeployment.

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


## TLS Settings

To allow TLS connections to Kafka for the consumer and produce use the following environment variables to the service orchestator section:

  * Set  KAFKA_SECURITY_PROTOCOL to "ssl"
  * If you have the values for keys and certificates use:
     * KAFKA_SSL_CA_CERT
     * KAFKA_SSL_CLIENT_CERT
     * KAFKA_SSL_CLIENT_KEY
  * If you have the file locations for the certificates use:
     * KAFKA_SSL_CA_CERT_FILE
     * KAFKA_SSL_CLIENT_CERT_FILE
     * KAFKA_SSL_CLIENT_KEY_FILE
  * If you key is password protected then add
     * KAFKA_SSL_CLIENT_KEY_PASS (optional)

An example spec that gets values from screts is shown below and comes from the [Kafka KEDA demo](../examples/kafka_keda.html).

```
   svcOrchSpec:
      env:
      - name: KAFKA_BROKER
        value: <nodepot>:9093
      - name: KAFKA_INPUT_TOPIC
        value: cifar10-rest-input
      - name: KAFKA_OUTPUT_TOPIC
        value: cifar10-rest-output
      - name: KAFKA_SECURITY_PROTOCOL
        value: ssl
      - name: KAFKA_SSL_CA_CERT
        valueFrom:
          secretKeyRef:
            name: seldon-cluster-ca-cert
            key: ca.crt
      - name: KAFKA_SSL_CLIENT_CERT
        valueFrom:
          secretKeyRef:
            name: seldon-user
            key: user.crt
      - name: KAFKA_SSL_CLIENT_KEY
        valueFrom:
          secretKeyRef:
            name: seldon-user
            key: user.key
      - name: KAFKA_SSL_CLIENT_KEY_PASS
        valueFrom:
          secretKeyRef:
            name: seldon-user
            key: user.password
```

## KEDA Scaling

KEDA can be used to scale Kafka SeldonDeployments by looking at the consumer lag. 

```
      kedaSpec:
        pollingInterval: 15
        minReplicaCount: 1
        maxReplicaCount: 2
        triggers:
        - type: kafka
          metadata:
            bootstrapServers: <nodeport>:9093
            consumerGroup: model.tfserving-cifar10.kafka
            lagThreshold: "50"
            topic: cifar10-rest-input
            offsetResetPolicy: latest
            #authMode: sasl_ssl (for latest KEDA - not released yet)
          authenticationRef:
            name: seldon-kafka-auth
```

In the above we:

 * define bootstrap servers for KEDA to connect to via `bootstrapServer`
 * define consumer group to monitor via `consumerGroup`
 * set the lag to scale up on via `lagThreshold`
 * monitor a particular topic via `topic`
 * define TLS authentication via a AuthenticanTrigger via `authenticationRef`

The authentication trigger we used for this was extracting the TLS details from secrets, e.g.

```
apiVersion: keda.sh/v1alpha1
kind: TriggerAuthentication
metadata:
  name: seldon-kafka-auth
  namespace: kafka
spec:
  secretTargetRef:
  - parameter: tls
    name: keda-enable-tls
    key: tls
  - parameter: ca
    name: seldon-cluster-ca-cert
    key: ca.crt
  - parameter: cert
    name: seldon-user
    key: user.crt
  - parameter: key
    name: seldon-user
    key: user.key
```

A worked example can be found [here](../examples/kafka_keda.html).

## Examples

 * [A worked example for a CIFAR10 image classifier is available](../examples/cifar10_kafka.html).
 * [A worked example for Kafka with KEDA scaling using TLS connections to Kafka](../examples/kafka_keda.html).
