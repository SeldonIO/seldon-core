# Service Orchestrator Engine

The service orchestrator is a component that is added to your inference graph to:

- Correctly manage the request/response paths described by your inference graph
- Expose Prometheus metrics
- Add meta data to the response

## Resource Requests/Limits for Service Orchestrator

You can set custom resource request and limits for this component by specifying them in a `svcOrchSpec` section in your Seldon Deployment. An example is shown below to set the engine cpu and memory requests:

```JSON
{
  "apiVersion": "machinelearning.seldon.io/v1alpha2",
  "kind": "SeldonDeployment",
  "metadata": {
    "name": "svcorch"
  },
  "spec": {
    "name": "resources",
    "predictors": [
      {
        "componentSpecs": [
          {
            "spec": {
              "containers": [
                {
                  "image": "seldonio/mock_classifier:1.0",
                  "name": "classifier"
                }
              ]
            }
          }
        ],
        "graph": {
          "children": [],
          "name": "classifier",
          "type": "MODEL",
          "endpoint": {
            "type": "REST"
          }
        },
        "svcOrchSpec": {
          "resources": {
            "requests": {
               "cpu": "1",
               "memory": "3Gi"
            }
          }
        },
        "name": "release-name",
        "replicas": 1
      }
    ]
  }
}

```

### Java Settings

The service orchestrator is a Java component. You can directly control its java settings as describe [here](../graph/annotations.html#service-orchestrator)

## Environment Variables for Service Orchestrator

You can manipulate some of the functionality of the service orchestrator by adding specific environment variables to the `svcOrchSpec` section.

- [Configure Jaeger Tracing Example](../graph/distributed-tracing.html)
- [Set logging level in service orchestrator engine](../analytics/log_level.html#setting-log-level-in-the-seldon-engine)

## Bypass Service Orchestrator (version >= 0.5.0, alpha feature)

If you are deploying a single model then for those wishing to minimize the latency and resource usage for their deployed model you can opt out of having the service orchestrator included. To do this add the annotation `seldon.io/no-engine: "true"` to the predictor. The predictor must contain just a single node graph. An example is shown below:

```YAML
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: noengine
spec:
  name: noeng
  predictors:
  - annotations:
      seldon.io/no-engine: "true"
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: noeng
    replicas: 1
```

In these cases the external API requests will be sent directly to your model. At present only the python wrapper (>=0.13-SNAPSHOT) has been modified to allow this.

Note no metrics or extra data will be added to the request so this would need to be done by your model itself if needed.

## Floating-point and integer numbers

We use [protobuf](https://developers.google.com/protocol-buffers/) to describe
the format of the input and output messages.
You can see the [reference for the SeldonMessage
object](../reference/apis/prediction.md) for more information about the actual
format.

As part of the options to specify the input request, you can use the `jsonData`
key to submit an arbitrary json.
To serialise the info in `jsonData`, we use the
[google.protobuf.Struct](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Struct)
and
[google.protobuf.Value](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#value)
types.
One of the caveats of these types is that they don't support integers.
Instead, they treat all numbers as floating-point numbers to align with the
JSON specification, where there is no distinction between both.
Therefore, when the service orchestrator parses the request, it will always
treat integers as floating-point numbers.
This behaviour can cause issues down the line if the nodes of the inference
graph expect integers.

To illustrate this problem, we can think of a payload such as:

```JSON
{
  "jsonData": {
    "vocabulary_length": 257,
    "threshold": 0.78,
    "sentence": "This is our input text"
  }
}
```

Because of how the `protobuf` types `google.protobuf.Struct` and
`google.protobuf.Value` work, the value of the `jsonData.vocabulary_length`
field will be parsed as a floating-point number `257.0` in the service
orchestrator.
By default, this would then get serialised and sent downstream as:

```JSON
{
  "jsonData": {
    "vocabulary_length": 257.0,
    "threshold": 0.78,
    "sentence": "This is our input text"
  }
}
```

The nodes of the inference graph would then parse the above as a floating-point
number, which could cause issues on any part that requires an integer input.

As a workaround, **the orchestrator omits empty decimal parts** when it
serialises the request before sending it to downstream nodes.
Going back to the example above, the orchestrator will serialise that input
payload as:

```JSON
{
  "jsonData": {
    "vocabulary_length": 257,
    "threshold": 0.78,
    "sentence": "This is our input text"
  }
}
```

Note that, if the decimal part is not empty the orchestrator will respect it.
