## Model with Engine Java Opts

```json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "name": "seldon-model"
    },
    "spec": {
        "name": "test-deployment",
	"annotations" : { "seldon.io/engine-java-opts" : "-Xmx1G" },
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