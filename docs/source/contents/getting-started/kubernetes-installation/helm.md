# Helm Install

We provide several Helm charts.

 * `seldon-core-v2-crds` : cluster wide install of custom resources
 * `seldon-core-v2-setup` : installation of core components
 * `seldon-core-v2-servers` : a default set of servers
 * `seldon-core-v2-certs` : a default set of certificates for TLS

The Helm charts can be found within the `k8s/helm-charts` folder and they are published [here](https://github.com/SeldonIO/helm-charts)

Assuming you have installed any ecosystem components: Jaeger, Prometheus, Kafka as discussed [here](./index.md) you can follow the
following steps.

Note that for Kafka follow the steps discussed [here](kafka.md) 

## Add Seldon Core v2 Charts

```bash
helm repo add seldon-charts https://seldonio.github.io/helm-charts
helm repo update seldon-charts
```

## Install the CRDs

```bash
helm install seldon-core-v2-crds  seldon-charts/seldon-core-v2-crds
```

## Install the Seldon Core V2 Components

You can install into any namespace. For illustration we will use `seldon-mesh`. By default Seldon runs in namespaced mode and needs only namespaced Roles. Model, Pipeline, Experiment Resources will need to be created in the chosen namespace.

```bash
kubectl create namespace seldon-mesh
```

```bash
helm install seldon-core-v2  seldon-charts/seldon-core-v2-setup --namespace seldon-mesh
```

## Install the Default Seldon Core V2 Servers

```bash
helm install seldon-v2-servers seldon-charts/seldon-core-v2-servers --namespace seldon-mesh
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