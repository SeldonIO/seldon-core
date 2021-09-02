# Payload Logging

Logging of request and response payloads from your Seldon Deployment can be accomplished by adding a logging section to any part of the Seldon deployment graph. An example is shown below:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.3
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
      logger:
        url: http://mylogging-endpoint
        mode: all
    name: example
    replicas: 1

```

The logging for the top level requets response is provided by:

```yaml
      logger:
        url: http://mylogging-endpoint
        mode: all
```

In this example both request and response payloads as specified by the `mode` attribute are sent as [CloudEvents](https://cloudevents.io/) to the url `http://mylogging-endpoint`.

The specification is:

 * url: Any url. Optional. If not provided then it will default to the default knative borker in the namespace of the Seldon Deployment.
 * mode: Either `request`, `response` or `all`

## Logging direct to Kafka

You can log requests directly to Kafka as an alternative to logging via CloudEvents by adding appropriate environment variables to the `svcOrchSpec`. An example is shown below:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: cifar10
  namespace: seldon
spec:
  name: resnet32
  predictors:
  - graph:
      implementation: TRITON_SERVER
      logger:
        mode: all
      modelUri: gs://seldon-models/triton/tf_cifar10
      name: cifar10
    name: default
    svcOrchSpec:
      env:
      - name: LOGGER_KAFKA_BROKER
        value: seldon-kafka-plain-0.kafka:9092
      - name: LOGGER_KAFKA_TOPIC
        value: seldon
    replicas: 1
  protocol: kfserving

```

The two required environment variables are:

 * LOGGER_KAFKA_BROKER : The Kafka Broker service endpoint.
 * LOGGER_KAFKA_TOPIC : The kafka Topic to log the requests.

Follow a [benchmarking notebook for CIFAR10 image payload logging showing 3K predictions per second with Triton Inference Server](../examples/kafka_logger.html).

## Setting Global Default

If you don't want to set up the custom logger every time, you are able to set it with `executor.requestLogger.defaultEndpoint` in the Helm Chart Variable as outlined in the [helm chart advanced settings section](../reference/helm.rst). 

This can simply specify a URL to call. In the usual kubernetes fashion, if a service name is provided then it is assumed to be in the current namespace unless there it is followed by `.<namespace>`, giving the namespace name. 

You will still want to make sure the model is deployed with a specification on what requests will be logged, i.e. all, request or response (as outlined above).


### Example Notebook

You can try out an [example notebook with logging](../examples/payload_logging.html)

