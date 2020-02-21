## Model with REST and gRPC Settings

```json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "name": "seldon-model"
    },
    "spec": {
        "annotations": {
	    "seldon.io/grpc-max-message-size":"10000000",
	    "seldon.io/rest-timeout":"100000",	    
	    "seldon.io/grpc-timeout":"100000"
        },
        "name": "test-deployment",
        "predictors": [
            {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "image": "seldonio/mock_classifier_grpc:1.0",
                                "name": "classifier",
                            }
                        ]
                    }
                }],
                "graph": {
                    "children": [],
                    "name": "classifier",
                    "endpoint": {
			"type" : "GRPC"
		    },
                    "type": "MODEL"
                },
                "name": "grpc-size",
                "replicas": 1
            }
        ]
    }
}

```