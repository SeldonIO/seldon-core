The temaplate json files are generated from seldon-core helm charts.

The template_0.2.json is generated using:

```
git clone --branch release-0.2 git@github.com:SeldonIO/seldon-core.git seldon-core-release-0.2
helm template --set ambassador.enabled=true --set single_namespace=true seldon-core-release-0.2/helm-charts/seldon-core > template_0.2_single_namespace.yaml
helm template --set ambassador.enabled=true --set single_namespace=false seldon-core-release-0.2/helm-charts/seldon-core > template_0.2_cluster_wide.yaml
kubectl convert -f template_0.2_single_namespace.yaml -o json > template_0.2_single_namespace.json
kubectl convert -f template_0.2_cluster_wide.yaml -o json > template_0.2_cluster_wide.json
rm template_0.2*.yaml
```

The template_0.1.json is generated using:

```
git clone --branch release-0.1 git@github.com:SeldonIO/seldon-core.git seldon-core-release-0.1
helm template --set ambassador.enabled=true seldon-core-release-0.1/helm-charts/seldon-core > template_0.1.yaml
kubectl convert -f template_0.1.yaml -o json > template_0.1.json
rm template_0.1.yaml
```