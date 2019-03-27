# GRPC Load Balancing with Ambassador

```json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "name": "seldon-model"
    },
    "spec": {
	"annotations": {
	    "seldon.io/headless-svc":"true"
	},
        "name": "test-deployment",
        "predictors": [
            {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "image": "seldonio/mock_classifier_grpc:1.0",
                                "name": "classifier"
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
                "name": "example",
                "replicas": 4
            }
        ]
    }
}

```