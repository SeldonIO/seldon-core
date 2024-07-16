# Service Meshes

The Seldon models and pipelines are exposed via a single service endpoint in the install namespace called `seldon-mesh`. All models, pipelines and experiments can be reached via this single Service endpoint by setting appropriate headers on the inference REST/gRPC request. By this means Seldon is agnostic to any service mesh you may wish to use in your organisation. We provide some example integrations for some example service meshes below (alphabetical order):

 * [Ambassador](./ambassador/index.md)
 * [Istio](./istio/index.md)
 * [Traefik](./traefik/index.md)


We welcome help to extend these to other service meshes.

```{toctree}
:maxdepth: 1
:hidden:

ambassador/index.md
istio/index.md
traefik/index.md
```

