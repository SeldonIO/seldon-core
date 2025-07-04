# Upgrading

## Upgrading from 2.6 - 2.7

All pods provisioned through the operator i.e. `SeldonRuntime` and `Server` resources now have the label `app.kubernetes.io/name` for identifying the pods. 

Previously, the labelling has been inconsistent across different versions of Seldon Core 2, with mixture of `app` and `app.kubernetes.io/name` used.

If using the Prometheus operator ("Kube Prometheus"), please apply the v2.7.0 manifests for Seldon Core 2 according to the [metrics documentation](../kubernetes/metrics/index.md).

Note that these manifests need to be adjusted to discover metrics endpoints based on the existing setup.

If previous pod monitors had `namespaceSelector` fields set, these should be copied over and applied to the new manifests.

If namespaces do not matter, cluster-wide metrics endpoint discovery can be setup by modifying the `namespaceSelector` field in the pod monitors:
```yaml
spec:
  namespaceSelector:
    any: true
```

## Upgrading from 2.5 - 2.6

Release 2.6 brings with it new custom resources `SeldonConfig` and `SeldonRuntime`, which provide a new way to install Seldon Core 2 in Kubernetes. Upgrading in the same namespace will cause downtime while the pods are being recreated. Alternatively  users can have an external service mesh or other means to be used over multiple namespaces to bring up the system in a new namespace and redeploy models before switch traffic between them.

If the new 2.6 charts are used to upgrade in an existing namespace models will eventually be redeloyed but there will be service downtime as the core components are redeployed.

