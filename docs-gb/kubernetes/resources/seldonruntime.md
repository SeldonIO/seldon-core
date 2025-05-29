---
description: Learn about SeldonRuntime, a Kubernetes resource for creating and managing Seldon Core instances in specific namespaces with configurable settings.
---

# SeldonRuntime Resource

The SeldonRuntime resource is used to create an instance of Seldon installed in a particular namespace.

```go
type SeldonRuntimeSpec struct {
	SeldonConfig string              `json:"seldonConfig"`
	Overrides    []*OverrideSpec     `json:"overrides,omitempty"`
	Config       SeldonConfiguration `json:"config,omitempty"`
	// +Optional
	// If set then when the referenced SeldonConfig changes we will NOT update the SeldonRuntime immediately.
	// Explicit changes to the SeldonRuntime itself will force a reconcile though
	DisableAutoUpdate bool `json:"disableAutoUpdate,omitempty"`
}

type OverrideSpec struct {
	Name        string         `json:"name"`
	Disable     bool           `json:"disable,omitempty"`
	Replicas    *int32         `json:"replicas,omitempty"`
	ServiceType v1.ServiceType `json:"serviceType,omitempty"`
	PodSpec     *PodSpec       `json:"podSpec,omitempty"`
}
```

For the definition of `SeldonConfiguration` above see the [SeldonConfig resource](seldonconfig.md).

The specification above contains overrides for the chosen `SeldonConfig`.
To override the `PodSpec` for a given component, the `overrides` field needs to specify the component
name and the `PodSpec` needs to specify the container name, along with fields to override.

For instance, the following overrides the resource limits for `cpu` and `memory` in the `hodometer`
component in the `seldon-mesh` namespace, while using values specified in the `seldonConfig` elsewhere
(e.g. `default`).

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: SeldonRuntime
metadata:
  name: seldon
  namespace: seldon-mesh
spec:
  overrides:
  - name: hodometer
    podSpec:
      containers:
      - name: hodometer
        resources:
          limits:
            memory: 64Mi
            cpu: 20m
  seldonConfig: default
```

As a minimal use you should just define the `SeldonConfig` to use as a base for this install, for
example to install in the `seldon-mesh` namespace with the `SeldonConfig` named `default`:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: SeldonRuntime
metadata:
  name: seldon
  namespace: seldon-mesh
spec:
  seldonConfig: default
```

The helm chart `seldon-core-v2-runtime` allows easy creation of this resource and associated default
Servers for an installation of Seldon in a particular namespace.

## SeldonConfig Update Propagation

When a [SeldonConfig](seldonconfig.md) resource changes any SeldonRuntime resources that
reference the changed SeldonConfig will also be updated immediately. If this behaviour is not desired
you can set `spec.disableAutoUpdate` in the SeldonRuntime resource for it not be be updated immediately
but only when it changes or any owned resource changes.
