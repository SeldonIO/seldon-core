seldon-single-model
===================
Chart to deploy a model in Seldon Core.

The chart bli bla blu


Current chart version is `0.2.0`

Source code can be found [here](https://github.com/SeldonIO/seldon-core)



## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| annotations | list | `[]` | Annotations applied to the deployment |
| hpa.enabled | bool | `false` | Whether to add an HPA spec to the deployment |
| hpa.maxReplicas | int | `5` | Maximum number of replicas for HPA |
| hpa.metrics | list | `[{"resource":{"name":"cpu","targetAverageUtilization":10},"type":"Resource"}]` | Metrics that autoscaler should check |
| hpa.minReplicas | int | `1` | Minimum number of replicas for HPA |
| labels | list | `[]` | Labels applied to the deployment |
| model.env | object | `{}` | Environment variables injected into the model's container |
| model.image | string | `""` | Docker image used by the model |
| model.resources | object | `{"requests":{"cpu":"500m"}}` | Resource requests and limits for the model's container |
| replicas | int | `1` | Number of replicas for the predictor |
