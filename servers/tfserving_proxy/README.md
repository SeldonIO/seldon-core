# TensorFlow Serving Proxy

The TensorFlow Serving Proxy provides a proxy to forward Seldon prediction requests to a running [TensorFlow Serving](https://www.tensorflow.org/serving/) server.

## Configuration

The tensorflow proxy takes several parameters:

 | Parameter | Type | Value | Example |
 |-----------|------|-------|---------|
 | rest_endpoint | STRING | URL of server HTTP endpoint | http://0.0.0.0:8000 |
 | grpc_endpoint | STRING | host and port of gRPC endpoint | 0.0.0.0:8001 |
 | model_name | STRING | model name | mnist-model |
 | signature_name | STRING | model signature name | predict_images |
 | model_input | STRING | model input name | images |
 | model_output | STRING | model output name | scores |

An example resource with the proxy and a Tensorflow Serving server is shown below.


```json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "labels": {
            "app": "seldon"
        },
        "name": "tfserving-mnist",
	"namespace": "seldon"	
    },
    "spec": {
        "name": "tf-mnist",
        "predictors": [
            {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "image": "seldonio/tfserving-proxy:0.1",
                                "name": "tfserving-proxy"
                            },
			    {
				"args": [
				    "/usr/bin/tensorflow_model_server",
				    "--port=8000",
				    "--model_name=mnist-model",
				    "--model_base_path=gs://seldon-models/tfserving/mnist-model"
				],
				"image": "gcr.io/kubeflow-images-public/tensorflow-serving-1.7:v20180604-0da89b8a",
				"name": "mnist-model",
				"ports": [
				    {
					"containerPort": 8000,
					"protocol": "TCP"
				    }
				],
				"resources": {
				    "limits": {
					"cpu": "4",
					"memory": "4Gi"
				    },
				    "requests": {
					"cpu": "1",
					"memory": "1Gi"
				    }
				},
				"securityContext": {
				    "runAsUser": 1000
				}
			    }
			],
			"terminationGracePeriodSeconds": 1
		    }
		}],
                "graph": {
		    "name": "tfserving-proxy",
		    "endpoint": { "type" : "REST" },
		    "type": "MODEL",
		    "children": [],
		    "parameters":
		    [
			{
			    "name":"grpc_endpoint",
			    "type":"STRING",
			    "value":"localhost:8000"
			},
			{
			    "name":"model_name",
			    "type":"STRING",
			    "value":"mnist-model"
			},
			{
			    "name":"model_output",
			    "type":"STRING",
			    "value":"scores"
			},
			{
			    "name":"model_input",
			    "type":"STRING",
			    "value":"images"
			},
			{
			    "name":"signature_name",
			    "type":"STRING",
			    "value":"predict_images"
			}
		    ]
		},
                "name": "mnist-tfserving",
                "replicas": 1
            }
        ]
    }
}
```

Examples:

 * [MNIST with TensorFlow Serving Proxy](https://github.com/SeldonIO/seldon-core/blob/master/examples/models/tfserving-mnist/tfserving-mnist.ipynb).
