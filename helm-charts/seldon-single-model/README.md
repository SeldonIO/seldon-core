# seldon-single-model

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square)

Chart to deploy a model in Seldon Core.

The chart bli bla blu

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```shell
helm repo add seldonio https://storage.googleapis.com/seldon-charts
```

Once that's done, you should be able to generate your `SeldonDeployment`
resources as:

```shell
helm template $MY_MODEL_NAME seldonio/seldon-single-model --namespace $MODELS_NAMESPACE
```

Note that you can also install / deploy the chart directly to your cluster using:

```shell
helm install $MY_MODEL_NAME seldonio/seldon-single-model --namespace $MODELS_NAMESPACE
```

**Homepage:** <https://github.com/SeldonIO/seldon-core>

## Source Code

* <https://github.com/SeldonIO/seldon-core>

## Values

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
