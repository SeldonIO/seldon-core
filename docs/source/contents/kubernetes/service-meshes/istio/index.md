# Istio

[Istio](https://istio.io/) provides a service mesh and ingress solution.

We will run through some examples as shown in the notebook `service-meshes/istio/istio.ipynb`

## Single Model

 * A Seldon Iris Model
 * An istio Gateway
 * An instio VirtualService to expose REST and gRPC

```{literalinclude} ../../../../../../service-meshes/istio/static/single-model.yaml 
:language: yaml
```

## Traffic Split

 * Two Iris Models
 * An istio Gateway
 * An istio VirtualService with traffic split

```{literalinclude} ../../../../../../service-meshes/istio/static/traffic-split.yaml 
:language: yaml
```

```{include} ../../../../../../service-meshes/istio/README.md
```
