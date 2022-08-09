# Helm Install

We provide two Helm charts.

 * `seldon-core-v2-crds` : cluster wide install of custom resources
 * `seldon-core-v2-setup` : installation of core components
 * `seldon-core-v2-servers` : a default set of servers

The Helm charts can be found within the `k8s/helm-charts` folder.

Assuming you have installed any ecosystem components: Jaeger, Prometheus, Kafka as discussed [here](./index.md) you can follow the
following steps.

## Install the CRDs

```bash
helm install seldon-core-v2-crds  k8s/helm-charts/seldon-core-v2-crds
```

## Install the Seldon Core V2 Components

You can install into any namespace. For illustration we will use `seldon-mesh`. By default Seldon runs in namespaced mode and needs only namespaced Roles. Model, Pipeline, Experiment Resources will need to be created in the chosen namespace.

```bash
kubectl create namespace seldon-mesh
```

```bash
helm install seldon-core-v2  k8s/helm-charts/seldon-core-v2-setup --namespace seldon-mesh
```

## Install the Default Seldon Core V2 Servers

```bash
helm install seldon-v2-servers k8s/helm-charts/seldon-core-v2-servers --namespace seldon-mesh
```

## Uninstall

Remove any models, pipelines that are running. 

Remove the servers:

```bash
helm uninstall seldon-core-v2-servers  --namespace seldon-mesh
```
Remove the core components:

```bash
helm uninstall seldon-core-v2  --namespace seldon-mesh
```

Remove the CRDs

```bash
helm uninstall seldon-core-v2-crds
```