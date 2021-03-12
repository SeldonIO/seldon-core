# Profiling/Debugging Service Orchestrator

# YourKit Profiler

  1. Create debug image using Makefile. Ensure you use current version of engine.
  1. Launch Seldon Core with this image as the engine image using, e.g.
     ```
     !helm install seldon-core ../helm-charts/seldon-core --namespace seldon  --set ambassador.enabled=true --set engine.image.name=seldonio/engine-debug:0.2.6-SNAPSHOT
     ```
  1. Check logs of instance to get port and port forward to it, e.g.
     ```
     kubectl port-forward mymodel-mymodel-7cd068f-54b4bfd4b5-njkrq  -n seldon 10001:10001
     ```
  1. Attach to JVM using [YourKit Java Profiler](https://www.ej-technologies.com/products/jprofiler/overview.html).

# JVisualVM

  1. Port forward to debug port which is added as default JAVA OPTS, e.g.
     ```
     kubectl port-forward mnist-classifier-mnist-classifier-svc-orch-5c49b566d7-bz7bl -n seldon 9090:9090
     ```
  1. Run [visualvm](https://visualvm.github.io/) and attach to process in list


# Remote Interactive Debug

  1. Annotate Seldon Deployment with ```"seldon.io/engine-java-opts" : "-agentlib:jdwp=transport=dt_socket,address=9090,server=y,suspend=n"```
  1. Apply SeldonDeployment
  1. Port forward the port used above
  1. From Eclipse create Remote Debug session using 127.0.0.1 and port above.

Example deployment:

```
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "labels": {
            "app": "seldon"
        },
        "name": "seldon-model"
    },
    "spec": {
        "name": "test-deployment",
        "predictors": [
            {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "image": "seldonio/mock_classifier:1.0",
                                "imagePullPolicy": "IfNotPresent",
                                "name": "classifier",
                                "resources": {
                                    "requests": {
                                        "memory": "1Mi"
                                    }
                                }
                            }
                        ],
                        "terminationGracePeriodSeconds": 1
                    }
                }],
                "graph": {
                    "children": [],
                    "name": "classifier",
                    "endpoint": {
			"type" : "REST"
		    },
                    "type": "MODEL"
                },
                "name": "example",
                "replicas": 1,
		"labels": {
		    "version" : "v1"
		},
		"annotations" : { "seldon.io/engine-java-opts" : "-agentlib:jdwp=transport=dt_socket,address=9090,server=y,suspend=n" }
            }
        ]
    }
}
```

# Local Interactive Debug

 1. Change ```engine/deploymentdef.json``` as needed. This is default Seldon Deployment that will be run.
 1. Run Engine in Eclipse and debug

To Run with local Model, for example:

Use mean_classifier from examples and run locally:

 ```
 PREDICTIVE_UNIT_SERVICE_PORT=5001 seldon-core-microservice MeanClassifier REST --service-type MODEL
 ```

Change ```engine/deploymentdef.json``` to:

```
           {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "image": "seldonio/mock_classifier_rest:1.0",
                                "imagePullPolicy": "IfNotPresent",
                                "name": "classifier",
                                "resources": {
                                    "requests": {
                                        "memory": "1Mi"
                                    }
                                }
                            }
                        ],
                        "terminationGracePeriodSeconds": 1
                    }
                }],
                "graph": {
                    "children": [],
                    "name": "classifier",
                    "endpoint": {
			"type" : "REST",
			"service_host" : "0.0.0.0",
			"service_port" : 5001
		    },
                    "type": "MODEL"
                },
                "name": "example",
                "replicas": 1,
		"labels": {
		    "version" : "v1"
		}
            }
```

Send requests:

```
curl 0.0.0.0:8081/api/v0.1/predictions -d '{"data":{"names":["a","b"],"ndarray":[[1,2]]}}}' -H "Content-Type: application/json"
```
