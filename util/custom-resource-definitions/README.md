# Utilities for creating validation template for CRD

## Pod Template Spec
run

```
make clean pod-template-spec-validation.tpl
```

## Horizontal Pod Autoscaler Spec

run

```
make hpa-spec.json
```


## Using the templates

To include the template you will need to manually include the output in

 * Helm chart templates
 * Ksonnet files




