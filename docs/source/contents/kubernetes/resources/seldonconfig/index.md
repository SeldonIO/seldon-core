# Seldon Config

```{note}
This section is for advanced usage where you want to define how seldon is installed in each namespace.
```

The SeldonConfig resource defines the core installation components installed by Seldon. If you wish to install Seldon, you can use the [SeldonRuntime](../seldonruntime/index.md) resource which allows easy overriding of some parts defined in this specification. In general, we advise core DevOps to use the default SeldonConfig or customize it for their usage. Individual installation of Seldon can then use the SeldonRuntime with a few overrides for special customisation needed in that namespace.

The specification contains core PodSpecs for each core component and a section for general configuration including the ConfigMaps that are created for the Agent (rclone defaults), Kafka and Tracing (open telemetry).


```{literalinclude} ../../../../../../operator/apis/mlops/v1alpha1/seldonconfig_types.go
:language: golang
:start-after: // SeldonConfigSpec
:end-before: // SeldonConfigStatus
```
Some of these values can be overridden on a per namespace basis via the SeldonRuntime resource.

The default configuration is shown below.


```{literalinclude} ../../../../../../operator/config/seldonconfigs/default.yaml
:language: yaml
```

