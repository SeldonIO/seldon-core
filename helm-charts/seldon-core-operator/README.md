# seldon-core-operator

![Version: 1.2.2-dev](https://img.shields.io/badge/Version-1.2.2-dev-informational?style=flat-square)

Seldon Core CRD and controller helm chart for Kubernetes.

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```shell
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

You can now deploy the chart as:

```shell
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
| credentials.gcs.gcsCredentialFileName | string | `"gcloud-application-credentials.json"` |  |
| credentials.s3.s3AccessKeyIDName | string | `"awsAccessKeyID"` |  |
| credentials.s3.s3SecretAccessKeyName | string | `"awsSecretAccessKey"` |  |
| defaultUserID | string | `"8888"` |  |
| engine.grpc.port | int | `5001` |  |
| engine.image.pullPolicy | string | `"IfNotPresent"` |  |
| engine.image.registry | string | `"docker.io"` |  |
| engine.image.repository | string | `"seldonio/engine"` |  |
| engine.image.tag | string | `"1.2.2-dev"` |  |
| engine.logMessagesExternally | bool | `false` |  |
| engine.port | int | `8000` |  |
| engine.prometheus.path | string | `"/prometheus"` |  |
| engine.serviceAccount.name | string | `"default"` |  |
| engine.user | int | `8888` |  |
| executor.enabled | bool | `true` |  |
| executor.image.pullPolicy | string | `"IfNotPresent"` |  |
| executor.image.registry | string | `"docker.io"` |  |
| executor.image.repository | string | `"seldonio/seldon-core-executor"` |  |
| executor.image.tag | string | `"1.2.2-dev"` |  |
| executor.metricsPortName | string | `"metrics"` |  |
| executor.port | int | `8000` |  |
| executor.prometheus.path | string | `"/prometheus"` |  |
| executor.requestLogger.defaultEndpoint | string | `"http://default-broker"` |  |
| executor.serviceAccount.name | string | `"default"` |  |
| executor.user | int | `8888` |  |
| explainer.image | string | `"seldonio/alibiexplainer:1.2.2-dev"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.registry | string | `"docker.io"` |  |
| image.repository | string | `"seldonio/seldon-core-operator"` |  |
| image.tag | string | `"1.2.2-dev"` |  |
| istio.enabled | bool | `false` |  |
| istio.gateway | string | `"istio-system/seldon-gateway"` |  |
| istio.tlsMode | string | `""` |  |
| kubeflow | bool | `false` |  |
| manager.cpuLimit | string | `"500m"` |  |
| manager.cpuRequest | string | `"100m"` |  |
| manager.memoryLimit | string | `"300Mi"` |  |
| manager.memoryRequest | string | `"200Mi"` |  |
| managerCreateResources | bool | `false` |  |
| predictiveUnit.defaultEnvSecretRefName | string | `""` |  |
| predictiveUnit.metricsPortName | string | `"metrics"` |  |
| predictiveUnit.port | int | `9000` |  |
| predictor_servers.MLFLOW_SERVER.grpc.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.MLFLOW_SERVER.grpc.image | string | `"seldonio/mlflowserver_grpc"` |  |
| predictor_servers.MLFLOW_SERVER.rest.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.MLFLOW_SERVER.rest.image | string | `"seldonio/mlflowserver_rest"` |  |
| predictor_servers.SKLEARN_SERVER.grpc.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.SKLEARN_SERVER.grpc.image | string | `"seldonio/sklearnserver_grpc"` |  |
| predictor_servers.SKLEARN_SERVER.rest.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.SKLEARN_SERVER.rest.image | string | `"seldonio/sklearnserver_rest"` |  |
| predictor_servers.TENSORFLOW_SERVER.grpc.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.TENSORFLOW_SERVER.grpc.image | string | `"seldonio/tfserving-proxy_grpc"` |  |
| predictor_servers.TENSORFLOW_SERVER.rest.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.TENSORFLOW_SERVER.rest.image | string | `"seldonio/tfserving-proxy_rest"` |  |
| predictor_servers.TENSORFLOW_SERVER.tensorflow | bool | `true` |  |
| predictor_servers.TENSORFLOW_SERVER.tfImage | string | `"tensorflow/serving:2.1.0"` |  |
| predictor_servers.XGBOOST_SERVER.grpc.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.XGBOOST_SERVER.grpc.image | string | `"seldonio/xgboostserver_grpc"` |  |
| predictor_servers.XGBOOST_SERVER.rest.defaultImageVersion | string | `"1.2.2-dev"` |  |
| predictor_servers.XGBOOST_SERVER.rest.image | string | `"seldonio/xgboostserver_rest"` |  |
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
