# Server

```{note}
The default installation will provide two initial servers: one MLServer and one Triton. You only need to define additional servers for advanced use cases.
```

A Server defines an inference server onto which models will be placed for inference. By default on installation two server StatefulSets will be deployed one MlServer and one Triton. An example Server definition is shown below:

```{literalinclude} ../../../../../../operator/config/servers/mlserver.yaml
:language: yaml
```

The main requirement is a reference to a ServerConfig resource in this case `mlserver`.

## Detailed Specs

```{literalinclude} ../../../../../../operator/apis/mlops/v1alpha1/server_types.go
:language: golang
:start-after: // ServerSpec
:end-before: // ServerStatus
```

## Custom Servers

One can easily utilize a custom image with the existing ServerConfigs. For example, the following defines an MLServer server with a custom image:

```{literalinclude} ../../../../../../samples/servers/custom-mlserver.yaml
:language: yaml
```

This server can then be targeted by a particular model by specifying this server name when creating the model, for example:

```{literalinclude} ../../../../../../samples/models/iris-custom-server.yaml
:language: yaml
```

