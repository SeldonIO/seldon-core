# Autoscaling Seldon Deployments

To autoscale your Seldon Deployment resources you can add Horizontal Pod Template Specifications to the Pod Template Specifications you create. There are three steps:

  1. Ensure you have a resource request for the metric you want to scale on if it is a standard metric such as cpu or memory.
  1. Add a HPA Spec refering to this Deployment. (We presently support v1beta1 version of k8s HPA Metrics spec)

To illustrate this we have an example Seldon Deployment below:

```json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
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
                                        "cpu": "0.5"
                                    }
                                }
                            }
                        ],
                        "terminationGracePeriodSeconds": 1
                    },
		    "hpaSpec":
		    {
			"minReplicas": 1,
			"maxReplicas": 4,
			"metrics": 
			    [ {
				"type": "Resource",
				"resource": {
				    "name": "cpu",
				    "targetAverageUtilization": 10
				}
			    }]
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

In the above we can see, we added resource requests for the cpu:

```
"resources": {
    "requests": {
	"cpu": "0.5"
    }
}
```

We added an HPA spec referring to the PodTemplateSpec:

```
"hpaSpecs":[
    {
	"minReplicas": 1,
	"maxReplicas": 4,
	"metrics": 
	[ {
	    "type": "Resource",
	    "resource": {
		"name": "cpu",
		"targetAverageUtilization": 10
	    }
	}]
    }], 
```

For a worked example see [this notebook](../examples/models/autoscaling/autoscaling_example.ipynb).
