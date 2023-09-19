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

## Custom Capabilities
Servers can be defined with a `capabilities` field to indicate custom configurations (e.g. Python dependencies). For instance:

```{literalinclude} ../../../../samples/servers/custom-mlserver-capabilities.yaml
:language: yaml
```

These `capabilities` override the ones from the `serverConfig: mlserver`. A model that takes advantage of this is shown below:

```{literalinclude} ../../../../samples/models/iris-custom-requirements.yaml
:language: yaml
```

This above model will be matched with the previous custom server `mlserver-134`.

Servers can also be set up with the `extraCapabilities` that add to existing capabilities from the referenced ServerConfig. For instance:

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

The `capabilities` field takes precedence over the `extraCapabilities` field.

For some examples see [here](../examples/custom-servers.md).


## Autoscaling of Servers

Within docker we don't support this but for Kubernetes see [here](../kubernetes/autoscaling/index.md)


