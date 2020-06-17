# Seldon Core Analytics

Seldon core metrics containers are able to provide different log levels. 

By default the containers come out of the box with WARNING as default log level.

This can be changed into DEBUG, INFO, WARNING and ERROR by following the following instructions.

## Setting the environment variable

### Setting log level in a Python Wrapper

The change can be done by setting the `SELDON_LOG_LEVEL` environment variable in the respective container.

For example, to set it in each container running with the python wrapper, you would do it as follows by adding the environment variable SELDON_LOG_LEVEL to the containers running images wrapped by the python wrapper:

```
...
"spec": {
  "containers": [
      { 
          "name": "mymodel",
          "image": "x.y:123",
          "env": [
              {
                  "name": SELDON_LOG_LEVEL,
                  "value": DEBUG
              }
          ]
      }
  ]
}
...
```

Once this has been set, it's possible to use the log in your wrapper code as follows:

```
import logging

log = logging.getLogger()

log.debug(...)
```

### Setting log level in the Seldon Engine

In order to set the log level in the SeldonEngine this can be done by providing the env option to the svcOrchSpec, as follows:

```
"svcOrchSpec": {
    "env": [
        {
            "name": "SELDON_LOG_LEVEL",
            "value": "DEBUG"
        }
    ]
}
```

Here is a full example configuration file with a loglevel selection.

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
                "componentSpecs": [
                    {
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
                    }
                ],
                "graph": {
                    "children": [],
                    "endpoint": {
                        "type": "REST"
                    },
                    "name": "classifier",
                    "type": "MODEL"
                },
                "labels": {
                    "version": "v1"
                },
                "name": "example",
                "replicas": 1,
                "svcOrchSpec": {
                    "env": [
                        {
                            "name": "SELDON_LOG_LEVEL",
                            "value": "DEBUG"
                        }
                    ]
                }
            }
        ]
    }
}
```


