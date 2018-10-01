# V1Alpha2 Update

  * The ```PredictorSpec componentSpec``` is now ```componentSpecs``` and takes a list of ```PodTemplateSpecs``` allowing you to split your runtime graph into separate Kubernetes Deployments as needed. See the new [proto definition](./proto/seldon_deployment.proto). To update existing resources:
      * Change ```    "apiVersion": "machinelearning.seldon.io/v1alpha1"``` to ```    "apiVersion": "machinelearning.seldon.io/v1alpha2"```
      * Change ```componentSpec``` -> ```componentSpecs``` and enclose the existing ```PodTemplateSpec``` in a single element list

For example change:

```
{
    "apiVersion": "machinelearning.seldon.io/v1alpha1",
    "kind": "SeldonDeployment",
    "metadata": {
        "labels": {
            "app": "seldon"
        },
        "name": "seldon-deployment-example"
    },
    "spec": {
        "annotations": {
            "project_name": "FX Market Prediction",
            "deployment_version": "v1"
        },
        "name": "test-deployment",
        "oauth_key": "oauth-key",
        "oauth_secret": "oauth-secret",
        "predictors": [
            {
                "componentSpec": {
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
                        "terminationGracePeriodSeconds": 20
                    }
                },
                "graph": {
                    "children": [],
                    "name": "classifier",
                    "endpoint": {
			"type" : "REST"
		    },
                    "type": "MODEL"
                },
                "name": "fx-market-predictor",
                "replicas": 1,
		"annotations": {
		    "predictor_version" : "v1"
		}
            }
        ]
    }
}

```

to

```
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "labels": {
            "app": "seldon"
        },
        "name": "seldon-deployment-example"
    },
    "spec": {
        "annotations": {
            "project_name": "FX Market Prediction",
            "deployment_version": "v1"
        },
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
                        "terminationGracePeriodSeconds": 20
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
                "name": "fx-market-predictor",
                "replicas": 1,
		"annotations": {
		    "predictor_version" : "v1"
		}
            }
        ]
    }
}

```
