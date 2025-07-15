# Logging and log level

Out of the box, your Seldon deployments will be pre-configured to a sane set of
defaults when it comes to logging.
These settings involve both the logging level and the structure of the log
messages.

These settings can be changed on a per-component basis.

## Log level

By default, all the components in your Seldon deployment will come out of the
box with `INFO` as the default log level.

To change the log level you can use the `SELDON_LOG_LEVEL` environment
variable.
In general, this variable can be set to the following log levels (from more to
less verbose):

- `DEBUG`
- `INFO`
- `WARNING` 
- `ERROR`

### Python inference servers

.. Note:: 
   Setting the ``SELDON_LOG_LEVEL`` to ``WARNING`` and above in the Python
   wrapper will disable the server's access logs, which are considered
   ``INFO``-level logs.

When using the [Python wrapper](../python/index) (including the
[MLflow](../servers/mlflow), [SKLearn](../servers/sklearn) and
[XGBoost](../servers/xgboost) pre-package servers), you can control the log
level using the `SELDON_LOG_LEVEL` environment variable.
Note that the `SELDON_LOG_LEVEL` variable has to be set in the **respective
container** within your inference graph.

For example, to set it in each container running with the python wrapper, you
would do it as follows by adding the environment variable `SELDON_LOG_LEVEL` to
the containers running images wrapped by the python wrapper:

```javascript
"spec": {
  // ...
  "predictors": [
    {
      "componentSpecs": [
        {
          "spec": {
            "containers": [
                { 
                    "name": "mymodel",
                    "image": "x.y:123",
                    "env": [
                        {
                            "name": "SELDON_LOG_LEVEL",
                            "value": "DEBUG"
                        }
                    ]
                }
            ]
          }
        }
      ]
    }
  ]
  // ...
}
```

Once this has been set, it's possible to use the log in your wrapper code as follows:

```python
import logging

log = logging.getLogger()
log.debug(...)
```

### Log level in the service orchestrator

To change the log level in the service orchestrator, you can set the
`SELDON_LOG_LEVEL`  environment variable on the `svcOrchSpec` section of the
`SeldonDeployment` CRD:

```javascript
"spec": {
  // ...
  "predictors": [
    {
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
  // ...
}
```

## Log format and sampling

By default, Seldon's service orchestrator and operator will serialise the log
messages as JSON and will enable log sampling.
This behaviour can be disabled by setting the `SELDON_DEBUG` variable to
`true`.
Note that this will **enable "debug mode"**, which can also have other side
effects.

For example, to change this on the service orchestrator, you would do:

```javascript
"spec": {
  // ...
  "predictors": [
    {
      "svcOrchSpec": {
          "env": [
              {
                  "name": "SELDON_DEBUG",
                  "value": "true"
              }
          ]
      }
    }
  ]
  // ...
}
```
