# Profiling Service Orchestrator

# YourKit Profiler

  1. Create debug image using Makefile. Ensure you use current version of engine.
  1. Launch Seldon Core with this image as the engine image using, e.g.
     ```
     !helm install ../helm-charts/seldon-core --name seldon-core --namespace seldon  --set ambassador.enabled=true --set engine.image.name=seldonio/engine-debug:0.2.6-SNAPSHOT
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


# Interactive Debug

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
        "oauth_key": "oauth-key",
        "oauth_secret": "oauth-secret",
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