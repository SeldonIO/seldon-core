# seldon-core-operator

![Version: 1.5.1](https://img.shields.io/static/v1?label=Version&message=1.5.1&color=informational&style=flat-square)

Seldon Core CRD and controller helm chart for Kubernetes.

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

Onca that's done, you should then be able to deploy the chart as:

```bash
kubectl create namespace seldon-system
helm install seldon-core-operator seldonio/seldon-core-operator --namespace seldon-system
```

## Source Code

* <https://github.com/SeldonIO/seldon-core>
* <https://github.com/SeldonIO/seldon-core/tree/master/helm-charts/seldon-core-operator>
* <https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| ambassador.enabled | bool | `true` |  |
| ambassador.singleNamespace | bool | `false` |  |
| certManager.enabled | bool | `false` |  |
| controllerId | string | `""` |  |
| crd.create | bool | `true` |  |
| crd.forceV1 | bool | `false` |  |
| crd.forceV1beta1 | bool | `false` |  |
| credentials.gcs.gcsCredentialFileName | string | `"gcloud-application-credentials.json"` |  |
| credentials.s3.s3AccessKeyIDName | string | `"awsAccessKeyID"` |  |
| credentials.s3.s3SecretAccessKeyName | string | `"awsSecretAccessKey"` |  |
| defaultUserID | string | `"8888"` |  |
| engine.grpc.port | int | `5001` |  |
| engine.image.pullPolicy | string | `"IfNotPresent"` |  |
| engine.image.registry | string | `"docker.io"` |  |
| engine.image.repository | string | `"seldonio/engine"` |  |
| engine.image.tag | string | `"1.5.1"` |  |
| engine.logMessagesExternally | bool | `false` |  |
| engine.port | int | `8000` |  |
| engine.prometheus.path | string | `"/prometheus"` |  |
| engine.resources.cpuLimit | string | `"500m"` |  |
| engine.resources.cpuRequest | string | `"500m"` |  |
| engine.resources.memoryLimit | string | `"512Mi"` |  |
| engine.resources.memoryRequest | string | `"512Mi"` |  |
| engine.serviceAccount.name | string | `"default"` |  |
| engine.user | int | `8888` |  |
| executor.enabled | bool | `true` |  |
| executor.image.pullPolicy | string | `"IfNotPresent"` |  |
| executor.image.registry | string | `"docker.io"` |  |
| executor.image.repository | string | `"seldonio/seldon-core-executor"` |  |
| executor.image.tag | string | `"1.5.1"` |  |
| executor.metricsPortName | string | `"metrics"` |  |
| executor.port | int | `8000` |  |
| executor.prometheus.path | string | `"/prometheus"` |  |
| executor.requestLogger.defaultEndpoint | string | `"http://default-broker"` |  |
| executor.resources.cpuLimit | string | `"500m"` |  |
| executor.resources.cpuRequest | string | `"500m"` |  |
| executor.resources.memoryLimit | string | `"512Mi"` |  |
| executor.resources.memoryRequest | string | `"512Mi"` |  |
| executor.serviceAccount.name | string | `"default"` |  |
| executor.user | int | `8888` |  |
| explainer.image | string | `"seldonio/alibiexplainer:1.5.1"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.registry | string | `"docker.io"` |  |
| image.repository | string | `"seldonio/seldon-core-operator"` |  |
| image.tag | string | `"1.5.1"` |  |
| istio.enabled | bool | `false` |  |
| istio.gateway | string | `"istio-system/seldon-gateway"` |  |
| istio.tlsMode | string | `""` |  |
| keda.enabled | bool | `false` |  |
| kubeflow | bool | `false` |  |
| manager.cpuLimit | string | `"500m"` |  |
| manager.cpuRequest | string | `"100m"` |  |
| manager.memoryLimit | string | `"300Mi"` |  |
| manager.memoryRequest | string | `"200Mi"` |  |
| managerCreateResources | bool | `false` |  |
| predictiveUnit.defaultEnvSecretRefName | string | `""` |  |
| predictiveUnit.metricsPortName | string | `"metrics"` |  |
| predictiveUnit.port | int | `9000` |  |
| predictor_servers.MLFLOW_SERVER.protocols.seldon.defaultImageVersion | string | `"1.5.1"` |  |
| predictor_servers.MLFLOW_SERVER.protocols.seldon.image | string | `"seldonio/mlflowserver"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.kfserving.defaultImageVersion | string | `"0.1.1"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.kfserving.image | string | `"seldonio/mlserver"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.seldon.defaultImageVersion | string | `"1.5.1"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.seldon.image | string | `"seldonio/sklearnserver"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.seldon.defaultImageVersion | string | `"1.5.1"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.seldon.image | string | `"seldonio/tfserving-proxy"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.tensorflow.defaultImageVersion | string | `"2.1.0"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.tensorflow.image | string | `"tensorflow/serving"` |  |
| predictor_servers.TRITON_SERVER.protocols.kfserving.defaultImageVersion | string | `"20.08-py3"` |  |
| predictor_servers.TRITON_SERVER.protocols.kfserving.image | string | `"nvcr.io/nvidia/tritonserver"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.kfserving.defaultImageVersion | string | `"0.1.1"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.kfserving.image | string | `"seldonio/mlserver"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.seldon.defaultImageVersion | string | `"1.5.1"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.seldon.image | string | `"seldonio/xgboostserver"` |  |
| rbac.configmap.create | bool | `true` |  |
| rbac.create | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `"seldon-manager"` |  |
| singleNamespace | bool | `false` |  |
| storageInitializer.cpuLimit | string | `"1"` |  |
| storageInitializer.cpuRequest | string | `"100m"` |  |
| storageInitializer.image | string | `"gcr.io/kfserving/storage-initializer:v0.4.0"` |  |
| storageInitializer.memoryLimit | string | `"1Gi"` |  |
| storageInitializer.memoryRequest | string | `"100Mi"` |  |
| usageMetrics.enabled | bool | `false` |  |
| webhook.port | int | `443` |  |
