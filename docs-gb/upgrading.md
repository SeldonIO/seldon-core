# Upgrading

## Upgrading from 2.8 - 2.9 
Though there are no breaking changes between 2.8 and 2.9, there are some new functionalties offered that require changes to fields in our CRDs:
* In Core 2.9 you can now set `minReplicas` to enable [partial scheduling](models/scheduling.md#partial-scheduling) of Models. This means that users will no longer have to wait for the full set of desired replicas before loading models onto servers (e.g. when scaling up).
* We've also added a `spec.llm` field to the Model CRD . The field is used by the PromptRuntime in Seldon's [LLM Module](https://docs.seldon.ai/llm-module) to reference a LLM model. Only one of spec.llm and spec.explainer should be set at a given time. This allows the deployment of multiple "models" acting as prompt generators for the same LLM.
* Due to the introduction of Server-autoscaling, it is important to understand what type of autoscaling you want to leverage, and how that can be configured. Below are configuratation that help set autoscaling behaviour. All options here have corresponding command-line arguments that can be passed to seldon-scheduler when not using helm as the install method. The following helm values can be set
    * `autoscaling.autoscalingModelEnabled`, with corresponding cmd line arg: `--enable-model-autoscaling` (defaults to false): enable or disable native model autoscaling based on lag thresholds. Enabling this assumes that lag (number of inference requests "in-flight") is a representative metric based on which to scale your models in a way that makes efficient use of resources.
    * `autoscaling.autoscalingServerEnabled` with corresponding cmd line arg: `--enable-server-autoscaling` (defaults to "true"): enable to use native server autoscaling, where the number of server replicas is set according to the number of replicas required by the models loaded onto that server.
    * `autoscaling.serverPackingEnabled` with corresponding cmd line arg: `--server-packing-enabled` (experimental, defaults to "false"): enable server packing to try and reduce the number of server replicas on model scale-down.
    * `autoscaling.serverPackingPercentage` with corresponding cmd line arg: `--server-packing-percentage` (experimental, defaults to "0.0"): controls the percentage of model replica removals (due to model scale-down or deletion) that should trigger packing

## Upgrading from 2.7 - 2.8

Core 2.8 introduces several new fields in our CRDs:
* `statefulSetPersistentVolumeClaimRetentionPolicy` enables users to configure the cleaning of PVC on their **servers**. This field is set to **retain** as default.
* `Status.selector` was introduced as a mandatory field for **models** in 2.8.4 and made optional in 2.8.5. This field enables autoscaling with HPA.
* `PodSpec` in the `OverrideSpec` for **SeldonRuntimes** enables users to customize how Seldon Core 2 pods are created. In particular, this also allows for setting custom taints/tolerations, adding additional containers to our pods, configuring custom security settings.

These added fields do not result in breaking changes, apart from 2.8.4 which required the setting of the `Status.selector` upon upgrading. This field was however changed to optional in the subsequent 2.8.5 release. Updating the CRDs (e.g. via helm) will enable users to benefit from the associated functionality.

## Upgrading from 2.6 - 2.7

All pods provisioned through the operator i.e. `SeldonRuntime` and `Server` resources now have the
label `app.kubernetes.io/name` for identifying the pods.

Previously, the labelling has been inconsistent across different versions of Seldon Core 2, with
mixture of `app` and `app.kubernetes.io/name` used.

If using the Prometheus operator ("Kube Prometheus"), please apply the v2.7.0 manifests for Seldon Core 2
according to the [metrics documentation](/kubernetes/metrics.md).

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

If the new 2.6 charts are used to upgrade in an existing namespace models will eventually be redeloyed but there will be service downtime as the core components are redeployed.
