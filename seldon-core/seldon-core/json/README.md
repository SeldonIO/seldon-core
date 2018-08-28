The template.json is generated from the helm source chart at https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core

```
helm template --set ambassador.enabled=true ${SELDON_CORE_HOME}/helm-charts/seldon-core > template.yaml
kubectl convert -f template.yaml -o json > template.json
rm template.yaml
```