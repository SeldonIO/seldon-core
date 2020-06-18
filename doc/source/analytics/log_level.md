# Logging and log level

Seldon core metrics containers are able to provide different log levels. 
By default the containers come out of the box with `WARNING` as default log
level (or the equivalent less verbose settings).

The verbosity level can be increased to also log `DEBUG` and `INFO` messages by
following these instructions.

## Log level in Python inference servers

.. Note:: 
   Setting the ``SELDON_LOG_LEVEL`` to ``WARNING`` and above in the Python
   wrapper will disable the server's access logs.

When using the [Python wrapper](../python) (including the
[MLflow](../servers/mlflow), [SKLearn](../servers/sklearn) and
[XGBoost](../servers/xgboost) pre-package servers), you can control the log
level using the `SELDON_LOG_LEVEL` environment variable.
This variable can be set to `DEBUG`, `INFO`, `WARNING` or `ERROR` to adjust the
log level accordingly.
Note that this has to be set in the **respective container** within your
inference graph.

For example, to set it in each container running with the python wrapper, you
would do it as follows by adding the environment variable `SELDON_LOG_LEVEL` to
the containers running images wrapped by the python wrapper:

```jsonc
// ...
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
// ...
```

Once this has been set, it's possible to use the log in your wrapper code as follows:

```python
import logging

log = logging.getLogger()
log.debug(...)
```

## Log level in the service orchestrator

.. Note:: 
   When using the Go implementation, setting the ``SELDON_LOG_LEVEL`` to
   ``WARNING`` and above in the service orchestrator will also change the
   structure of the log messages, which will be logged as JSON, and will enable
   log sampling.

In order to set the log level in the Seldon engine this can be done by
providing the env option to the svcOrchSpec, as follows:

```json
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

```json
{    
    "apiVersion": "machinelearning.seldon.io/v1",
    "kind": "SeldonDeployment",
    "metadata": {
        "labels": {
            "app": "seldon"
        },
        "name": "seldon-model"
    },
    "spec": {
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

