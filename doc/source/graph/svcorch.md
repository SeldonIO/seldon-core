# Service Orchestrator Engine

The service orchestrator is a component that is added to your inference graph to:

 * Correctly manage the request/response paths described by your inference graph
 * Expose prometheus metrics
 * Add meta data to the response

## Resource Requests/Limits for Service Orchetsrator

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

 * [Configure Jaeger Tracing Example](../graph/distributed-tracing.html)
 * [Set logging level in service orchestrator engine](../analytics/log_level.html#setting-log-level-in-the-seldon-engine)

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
