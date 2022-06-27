# seldon-abtest

![Version: 0.2.0](https://img.shields.io/static/v1?label=Version&message=0.2.0&color=informational&style=flat-square)

Chart to deploy an AB test in Seldon Core. Allows you to split traffic between two models.

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

Once that's done, you should then be able to use the inference graph template as:

```bash
helm template $MY_MODEL_NAME seldonio/seldon-abtest --namespace $MODELS_NAMESPACE
```

Note that you can also deploy the inference graph directly to your cluster
using:

```bash
helm install $MY_MODEL_NAME seldonio/seldon-abtest --namespace $MODELS_NAMESPACE
```

**Homepage:** <https://github.com/SeldonIO/seldon-core>

## Source Code

* <https://github.com/SeldonIO/seldon-core>
* <https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-abtest>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| modela.image.name | string | `"seldonio/mock_classifier"` |  |
| modela.image.version | string | `"1.15.0-dev"` |  |
| modela.name | string | `"classifier-1"` |  |
| modelb.image.name | string | `"seldonio/mock_classifier"` |  |
| modelb.image.version | string | `"1.15.0-dev"` |  |
| modelb.name | string | `"classifier-2"` |  |
| predictor.name | string | `"default"` |  |
| replicas | int | `1` |  |
| separate_pods | bool | `true` |  |
| traffic_modela_percentage | float | `0.5` |  |
