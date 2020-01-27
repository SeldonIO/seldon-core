# Distributed Tracing

You can use Open Tracing to trace your API calls to Seldon Core.

## Install Jaeger

You will need to install Jaeger on your Kubernetes cluster. Follow their [documentation](https://www.jaegertracing.io/docs/1.16/operator/)

## Configuration

You will need to annotate your Seldon Deployment resource with environment variables to make tracing active and set the appropriate Jaeger configuration variables.

  * For the Seldon Service Orchestrator you will need to set the environment variables in the ```spec.predictors[].svcOrchSpec.env``` section. See the [Jaeger Java docs](https://github.com/jaegertracing/jaeger-client-java/tree/master/jaeger-core#configuration-via-environment) for available configuration variables.
  * For each Seldon component you run (e.g., model transformer etc.) you will need to add environment variables to the container section.


### Python Wrapper Configuration

Add an environment variable: TRACING with value 1 to activate tracing.

You can utilize the default configuration by simply providing the name of the Jaeger agent service by providing JAEGER_AGENT_HOST environment variable. Override default Jaeger agent port `5775` by setting JAEGER_AGENT_PORT environment variable.

To provide a custom configuration following the Jaeger Python configuration yaml defined [here](https://github.com/jaegertracing/jaeger-client-python) you can provide a configmap and the path to the YAML file in JAEGER_CONFIG_PATH environment variable.

An example is show below:

```json
{
    "apiVersion": "machinelearning.seldon.io/v1alpha2",
    "kind": "SeldonDeployment",
    "metadata": {
        "labels": {
            "app": "seldon"
        },
        "name": "tracing-example",
	"namespace": "seldon"	
    },
    "spec": {
        "name": "tracing-example",
        "oauth_key": "oauth-key",
        "oauth_secret": "oauth-secret",
        "predictors": [
            {
                "componentSpecs": [{
                    "spec": {
                        "containers": [
                            {
                                "name": "model1",				
                                "image": "seldonio/mock_classifier_rest:1.1",
				"env": [
				    {
					"name": "TRACING",
					"value": "1"
				    },
				    {
					"name": "JAEGER_CONFIG_PATH",
					"value": "/etc/tracing/config/tracing.yml"
				    }
				],
				"volumeMounts": [
				    {
					"mountPath": "/etc/tracing/config",
					"name": "tracing-config"
				    }
				]
                            }
			],
			"terminationGracePeriodSeconds": 1,
			"volumes": [
			    {
				"name": "tracing-config",
				"volumeSource" : {
				    "configMap": {
					"localObjectReference" :
					{
					    "name": "tracing-config"
					},
					"items": [
					    {
						"key": "tracing.yml",
						"path":  "tracing.yml"
					    }
					]
				    }
				}
			    }
			]
		    }
		}],
                "graph": {
		    "name": "model1",
		    "endpoint": { "type" : "REST" },
		    "type": "MODEL",
		    "children": [
		    ]
		},
                "name": "tracing",
                "replicas": 1,
		"svcOrchSpec" : {
		    "env": [
			{
			    "name": "TRACING",
			    "value": "1"
			},
			{
			    "name": "JAEGER_AGENT_HOST",
			    "value": "jaeger-agent"
			},
			{
			    "name": "JAEGER_AGENT_PORT",
			    "value": "5775"
			},
			{
			    "name": "JAEGER_SAMPLER_TYPE",
			    "value": "const"
			},
			{
			    "name": "JAEGER_SAMPLER_PARAM",
			    "value": "1"
			}
		    ]				
		}
            }
        ]
    }
}
```
        


## REST Example

![jaeger-ui-rest](./jaeger-ui-rest-example.png)

## gRPC Example

![jaeger-ui-rest](./jaeger-ui-grpc-example.png)


## Worked Example

[A fully worked template example](../examples/tracing.html) is provided.
