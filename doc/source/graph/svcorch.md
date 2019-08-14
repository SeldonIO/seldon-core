# Service Orchestrator Engine

The service orchestrator is a component that is added to your inference graph to:

 * Correctly manage the request/response paths described by your inference graph
 * Expose prometheus metrics


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