# Servers

By default Seldon installs two server farms using MLServer and Triton with 1 replica each. Models are scheduled onto servers based on the server's resources and whether the capabilities of the server matches the requirements specified in the Model request. For example:

```{literalinclude} ../../../../samples/models/sklearn-iris-gs.yaml
:language: yaml
```

This model specifies the requirement `sklearn`

There is a default capabilities for each server as follows:

* MLServer
  ```{literalinclude} ../../../../operator/config/serverconfigs/mlserver.yaml
  :language: yaml
  :start-after: SELDON_SERVER_CAPABILITIES
  :end-before: SELDON_OVERCOMMIT_PERCENTAGE
* Triton
  ```{literalinclude} ../../../../operator/config/serverconfigs/triton.yaml
  :language: yaml
  :start-after: SELDON_SERVER_CAPABILITIES
  :end-before: SELDON_OVERCOMMIT_PERCENTAGE
  ```

## Extra Capabilities
Servers can be set up with the `extraCapabilities` field to indicate custom configurations (e.g. Python dependencies). For instance:

```{literalinclude} ../../../../samples/servers/mlserver-extra-capabilities.yaml
:language: yaml
```
This server, `mlserver-extra`, inherits a default set of capabilities via `serverConfig: mlserver`.
These defaults are discussed above.
The `extraCapabilities` are appended to these to create a single list of capabilities for this server.

Models can then specify requirements to select a server that satisfies those requirements as follows.
```{literalinclude} ../../../../samples/models/extra-model-requirements.yaml
:language: yaml
```


## Autoscaling of Servers

Within docker we don't support this but for Kubernetes see [here](../kubernetes/autoscaling/index.md)


