# Helm Installation

We provide several Helm charts.

 * `seldon-core-v2-crds` : cluster wide install of custom resources.
 * `seldon-core-v2-setup` : installation of the manager to manage resources in the namespace or clusterwide. This also installs default **SeldonConfig** and **ServerConfig** resources which allow Runtimes and Servers to be installed easily on demand.
 * `seldon-core-v2-runtime` : this installs a **SeldonRuntime** custom resource which creates the core components in a namespace.
 * `seldon-core-v2-servers` : this installs **Server** custom resources which provide example core servers to load models.
 * `seldon-core-v2-certs` : a default set of certificates for TLS.

The Helm charts can be found within the `k8s/helm-charts` folder and they are published [here](https://github.com/SeldonIO/helm-charts)

Assuming you have installed any ecosystem components: Jaeger, Prometheus, Kafka as discussed [here](./index.md) you can follow the
following steps.

Note that for Kafka follow the steps discussed [here](../../kubernetes/kafka/index)

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

You can install into any namespace. For illustration we will use `seldon-mesh`. This will install the core manager which will handle the key [resources](../../kubernetes/resources/index.md)  used by Seldon including the SeldonRuntime and Server resources.

```bash
kubectl create namespace seldon-mesh
```

```bash
helm install seldon-core-v2  seldon-charts/seldon-core-v2-setup --namespace seldon-mesh
```

This will install the operator namespaced so it will only control resources in the provided namespace. To allow cluster wide usage add the `--set controller.clusterwide=true`, e.g.

```
helm install seldon-core-v2  seldon-charts/seldon-core-v2-setup --namespace seldon-mesh --set controller.clusterwide=true
```

Cluster wide operations will require ClusterRoles to be created so when deploying be aware your user will require the required permissions. With cluster wide operations you can create SeldonRuntimes in any namespace.

## Install the default Seldon Core V2 Runtime

```bash
helm install seldon-v2-runtime seldon-charts/seldon-core-v2-runtime --namespace seldon-mesh
```

This will install the core components in your desired namespace.

## Install example servers

To install some MLServer and Triton servers you can either create Server resources yourself or for initial testing you can use our example Helm chart seldon-core-v2-servers:

```bash
helm install seldon-v2-servers seldon-charts/seldon-core-v2-servers --namespace seldon-mesh
```

By default this will install 1 MLServer and 1 Triton in the desired namespace. This namespace should be the same namespace you installed a Seldon Core Runtime.

## Uninstall

Remove any models, pipelines that are running.

Remove the runtime:

```bash
helm uninstall seldon-core-v2-runtime  --namespace seldon-mesh
```
Remove the core components:

```bash
helm uninstall seldon-core-v2  --namespace seldon-mesh
```

Remove the CRDs

```bash
helm uninstall seldon-core-v2-crds
```
