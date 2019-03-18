## Model with Separate Service Orchestrator Pod

```json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "name": "seldon-model"
    },
    "spec": {
        "annotations": {
	    "seldon.io/engine-separate-pod": "true"
        },
        "name": "test-deployment",
        "predictors": [
            {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "image": "seldonio/mock_classifier:1.0",
                                "name": "classifier"
                            }
                        ]
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
                "replicas": 1
            }
        ]
    }
}

```