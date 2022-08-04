# Helm Install

We provide two Helm charts.

 * `seldon-core-v2-crds` : cluster wide install of custom resources
 * `seldon-core-v2-setup` : installation of core components

The Helm charts can be found within the `k8s/helm-charts` folder.

Assuming you have installed any ecosystem components: Jaeger, Prometheus, Kafka as discussed [here](./index.md) you can follow the
following steps.

## Install the CRDs

```bash
helm install seldon-core-v2-crds  k8s/helm-charts/seldon-core-v2-crds
```

## Install the Seldon Core V2 Components

```bash
kubectl create namespace seldon-mesh
```

```bash
helm install seldon-core-v2  k8s/helm-charts/seldon-core-v2-setup --namespace seldon-mesh
```

## Uninstall

Remove any models, pipelines that are running. Remove the Server components manually: e.g. mlserver and triton Server custom resources.

Remove the components:

```bash
helm uninstall seldon-core-v2  --namespace seldon-mesh
```

Remove the CRDs

```bash
helm uninstall seldon-core-v2-crds
```