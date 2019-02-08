# Autoscaling Seldon Deployments

To autoscale your Seldon Deployment resources you can add Horizontal Pod Template Specifications to the Pod Template Specifications you create. There are three steps:

  1. Give the PodTemplateSpec you want to autoscale a name by adding appropriate metadata.
  1. Ensure you have a resource request for the metric you want to scale on if it is a standard metric such as cpu or memory.
  1. Add a HorizontalPodAutoscalerSpec refering to this Deployment.

To illustrate this we have an example Seldon Deployment below:

```
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "name": "seldon-model"
    },
    "spec": {
        "name": "test-deployment",
        "predictors": [
            {
                "componentSpecs": [{
		    "metadata":{
			"name":"my-dep"
		    },
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
                    }
                }],
		"hpaSpecs":[
		    {
			"scaleTargetRef": {			    
			    "apiVersion": "extensions/v1beta1",
			    "kind": "Deployment",
			    "name": "my-dep"},
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

In the above we can see, we added meta data to give the PodTemplate a name:

```
"metadata":{
    "name":"my-dep"
},
```

We added resource requests for the cpu:
```
"resources": {
    "requests": {
	"cpu": "0.5"
    }
}
```

We added a HPA spec referring to the PodTemplateSpec:

```
"hpaSpecs":[
    {
	"scaleTargetRef": {			    
	    "apiVersion": "extensions/v1beta1",
	    "kind": "Deployment",
	    "name": "my-dep"},
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
