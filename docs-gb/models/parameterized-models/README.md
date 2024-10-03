# Parameterized Models

The Model specification allows parameters to be passed to the loaded model to allow customization. For example:

```yaml
# samples/models/choice1.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: choice-is-one
spec:
  storageUri: "gs://seldon-models/scv2/examples/pandasquery"
  requirements:
  - mlserver
  - python
  parameters:
  - name: query
    value: "choice == 1"
```

This capability is only available for MLServer custom model runtimes. The named keys and
values will be added to the model-settings.json file for the provided model in the
`parameters.extra` Dict. MLServer models are able to read these values in their `load` method.

## Example Parameterized Models

* [Pandas Query](./pandasquery.md)
