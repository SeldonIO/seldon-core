# Seldon Runtime

The SeldonRuntime resource is used to create an instance of Seldon installed in a particular namespace.

The specification contains overrides for the SeldonConfig chosen.

```{literalinclude} ../../../../../../operator/apis/mlops/v1alpha1/seldonruntime_types.go
:language: golang
:start-after: // SeldonRuntimeSpec
:end-before: // SeldonRuntimeStatus
```

For the definition of `SeldonConfiguration` above see the [SeldonConfig resource](../seldonconfig/index.md).

As a minimal use you should just define the SeldonConfig to use as a base for this install, for example to install in the seldon-mesh namespace with the "default" SeldonConfig:

```
apiVersion: mlops.seldon.io/v1alpha1
kind: SeldonRuntime
metadata:
  name: seldon
  namespace: seldon-mesh  
spec:
  seldonConfig: default
```

The helm chart `seldon-core-v2-runtime` allows easy creation of this resource and associated defualt Servers for an installation of Seldon in a particular namespace.

