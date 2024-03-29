# seldon-abtest

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square)

Chart to deploy an AB test in Seldon Core v1. Allows you to split traffic between two models.

**Homepage:** <https://github.com/SeldonIO/seldon-core>

## Source Code

* <https://github.com/SeldonIO/seldon-core>
* <https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-abtest>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| modela.image.name | string | `"seldonio/mock_classifier"` |  |
| modela.image.version | string | `"1.18.0"` |  |
| modela.name | string | `"classifier-1"` |  |
| modelb.image.name | string | `"seldonio/mock_classifier"` |  |
| modelb.image.version | string | `"1.18.0"` |  |
| modelb.name | string | `"classifier-2"` |  |
| predictor.name | string | `"default"` |  |
| replicas | int | `1` |  |
| separate_pods | bool | `true` |  |
| traffic_modela_percentage | float | `0.5` |  |
