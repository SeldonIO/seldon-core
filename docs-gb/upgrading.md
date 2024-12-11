# Upgrading

## Upgrading from 2.7 - 2.8

Core 2.8 introduces several new fields in our CRDs:
* `statefulSetPersistentVolumeClaimRetentionPolicy` enables users to configure the cleaning of PVC on their **servers**. This field is set to **retain** as default.
* `Status.selector` was introduced as a mandatory field for **models** in 2.8.4 and made optional in 2.8.5. This field enables autoscaling with HPA.
* `PodSpec` in the `OverrideSpec` for **SeldonRuntimes** enables users to customize how Seldon Core 2 pods are created. In particular, this also allows for setting custom taints/tolerations, adding additional containers to our pods, configuring custom security settings.

These added fields do not result in breaking changes, apart from 2.8.4 which required the setting of the `Status.selector` upon upgrading (this was removed as mandatory in the subsequent 2.8.5 release). Updating the CRDs (e.g. via helm) will enable users to benefit from the associated functionality.

## Upgrading from 2.6 - 2.7

All pods provisioned through the operator i.e. `SeldonRuntime` and `Server` resources now have the
label `app.kubernetes.io/name` for identifying the pods.

Previously, the labelling has been inconsistent across different versions of Seldon Core 2, with
mixture of `app` and `app.kubernetes.io/name` used.

If using the Prometheus operator ("Kube Prometheus"), please apply the v2.7.0 manifests for Seldon Core 2
according to the [metrics documentation](kubernetes/metrics.md).

Note that these manifests need to be adjusted to discover metrics endpoints based on the existing setup.

If previous pod monitors had `namespaceSelector` fields set, these should be copied over and applied
to the new manifests.

If namespaces do not matter, cluster-wide metrics endpoint discovery can be setup by modifying the
`namespaceSelector` field in the pod monitors:

```yaml
spec:
  namespaceSelector:
    any: true
```

## Upgrading from 2.5 - 2.6

Release 2.6 brings with it new custom resources `SeldonConfig` and `SeldonRuntime`, which provide
a new way to install Seldon Core 2 in Kubernetes. Upgrading in the same namespace will cause downtime
while the pods are being recreated. Alternatively  users can have an external service mesh or other
means to be used over multiple namespaces to bring up the system in a new namespace and redeploy models
before switch traffic between them.

If the new 2.6 charts are used to upgrade in an existing namespace models will eventually be redeloyed
but there will be service downtime as the core components are redeployed.
