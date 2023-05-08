# Custom Resource Patching

To allow kustomize to do a strategic patch operation on our Custom Resources we need to give it the open api spec otherwise it will default to a normal patch and overwrite.

 * Dowload openapi spec by installing the CRDs on a cluster and running `kustomize openapi fetch > crds-schema.json`
 * As we use a custom podSpec the extensions needed in the OpenApi spec are not added. Also maybe dependong on [this](https://github.com/kubernetes/kubernetes/issues/82942) . Edit the areas you want strategic merge by adding extension settings, e.g. for containers keyd on `name`:
   ```
    "x-kubernetes-patch-merge-key": "name",
    "x-kubernetes-patch-strategy": "merge",
   ```

