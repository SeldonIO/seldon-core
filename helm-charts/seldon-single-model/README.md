# seldon-single-model

![Version: 0.2.0](https://img.shields.io/static/v1?label=Version&message=0.2.0&color=informational&style=flat-square)

Chart to deploy a machine learning model in Seldon Core.

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

Once that's done, you should then be able to use the inference graph template as:

```bash
helm template $MY_MODEL_NAME seldonio/seldon-single-model --namespace $MODELS_NAMESPACE
```

Note that you can also deploy the inference graph directly to your cluster
using:

```bash
helm install $MY_MODEL_NAME seldonio/seldon-single-model --namespace $MODELS_NAMESPACE
```

**Homepage:** <https://github.com/SeldonIO/seldon-core>

## Source Code

* <https://github.com/SeldonIO/seldon-core>
* <https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-single-model>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| annotations | object | `{}` | Annotations applied to the deployment |
| apiVersion | string | `"machinelearning.seldon.io/v1"` | Version of the SeldonDeployment CRD |
| hpa.enabled | bool | `false` | Whether to add an HPA spec to the deployment |
| hpa.maxReplicas | int | `5` | Maximum number of replicas for HPA |
| hpa.metrics | list | `[{"resource":{"name":"cpu","targetAverageUtilization":10},"type":"Resource"}]` | Metrics that autoscaler should check |
| hpa.minReplicas | int | `1` | Minimum number of replicas for HPA |
| labels | object | `{}` | Labels applied to the deployment |
| model.env | object | `{"LOG_LEVEL":"INFO"}` | Environment variables injected into the model's container |
| model.envSecretRefName | string | `""` | The model secret name for enviroment variables |
| model.image | string | `""` | Docker image used by the model |
| model.implementation | string | `""` | Implementation of Prepackaged Model Server |
| model.logger.enabled | bool | `false` |  |
| model.logger.url | string | `""` |  |
| model.mlflow.xtype | string | `""` |  |
| model.resources | object | `{"requests":{"memory":"1Mi"}}` | Resource requests and limits for the model's container |
| model.uri | string | `""` | Model's URI for prepackaged model server |
| protocol | string | `"seldon"` |  |
| replicas | int | `1` | Number of replicas for the predictor |
