# Traefik

[Traefik](https://doc.traefik.io/) provides a service mesh and ingress solution.

We will run through some examples as shown in the notebook `service-meshes/traefik/traefik.ipynb`

## Single Model

 * A Seldon Iris Model
 * Traefik Service
 * Traefik IngressRoute
 * Traefik Middleware for adding a header

```{literalinclude} ../../../../../../service-meshes/traefik/static/single-model.yaml 
:language: yaml
```

## Traffic Split

```{warning}
Traffic splitting does not presently work due to this [issue](https://github.com/emissary-ingress/emissary/issues/4062). We recommend you use a Seldon Experiment instead.
```

```{include} ../../../../../../service-meshes/traefik/README.md
```
