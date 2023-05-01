# seldon-core-operator

![Version: 1.16.0](https://img.shields.io/static/v1?label=Version&message=1.16.0&color=informational&style=flat-square)

Seldon Core CRD and controller helm chart for Kubernetes.

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

Once that's done, you should then be able to deploy the chart as:

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
| ambassador.enabled | bool | `false` |  |
| ambassador.singleNamespace | bool | `false` |  |
| ambassador.version | string | `"v2"` |  |
| certManager.enabled | bool | `false` |  |
| controllerId | string | `""` |  |
| crd.annotations | object | `{}` |  |
| crd.create | bool | `true` |  |
| crd.forceV1 | bool | `false` |  |
| crd.forceV1beta1 | bool | `false` |  |
| credentials.gcs.gcsCredentialFileName | string | `"gcloud-application-credentials.json"` |  |
| credentials.s3.s3AccessKeyIDName | string | `"awsAccessKeyID"` |  |
| credentials.s3.s3SecretAccessKeyName | string | `"awsSecretAccessKey"` |  |
| defaultUserID | string | `"8888"` |  |
| executor.fullHealthChecks | bool | `false` |  |
| executor.image.pullPolicy | string | `"IfNotPresent"` |  |
| executor.image.registry | string | `"docker.io"` |  |
| executor.image.repository | string | `"seldonio/seldon-core-executor"` |  |
| executor.image.tag | string | `"1.16.0"` |  |
| executor.metricsPortName | string | `"metrics"` |  |
| executor.port | int | `8000` |  |
| executor.prometheus.path | string | `"/prometheus"` |  |
| executor.requestLogger.defaultEndpoint | string | `"http://default-broker"` |  |
| executor.requestLogger.workQueueSize | int | `10000` |  |
| executor.requestLogger.writeTimeoutMs | int | `2000` |  |
| executor.resources.cpuLimit | string | `"500m"` |  |
| executor.resources.cpuRequest | string | `"500m"` |  |
| executor.resources.memoryLimit | string | `"512Mi"` |  |
| executor.resources.memoryRequest | string | `"512Mi"` |  |
| executor.serviceAccount.name | string | `"default"` |  |
| executor.user | int | `8888` |  |
| explainer.image | string | `"seldonio/alibiexplainer:1.16.0"` |  |
| explainer.image_v2 | string | `"seldonio/mlserver:1.3.0-alibi-explain"` |  |
| hostNetwork | bool | `false` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.registry | string | `"docker.io"` |  |
| image.repository | string | `"seldonio/seldon-core-operator"` |  |
| image.tag | string | `"1.16.0"` |  |
| istio.enabled | bool | `false` |  |
| istio.gateway | string | `"istio-system/seldon-gateway"` |  |
| istio.tlsMode | string | `""` |  |
| keda.enabled | bool | `false` |  |
| kubeflow | bool | `false` |  |
| manager.annotations | object | `{}` |  |
| manager.containerSecurityContext | object | `{}` |  |
| manager.cpuLimit | string | `"500m"` |  |
| manager.cpuRequest | string | `"100m"` |  |
| manager.deploymentNameAsPrefix | bool | `false` |  |
| manager.leaderElectionID | string | `"a33bd623.machinelearning.seldon.io"` |  |
| manager.leaderElectionLeaseDurationSecs | int | `15` |  |
| manager.leaderElectionRenewDeadlineSecs | int | `10` |  |
| manager.leaderElectionResourceLock | string | `nil` |  |
| manager.leaderElectionRetryPeriodSecs | int | `2` |  |
| manager.logLevel | string | `"INFO"` |  |
| manager.memoryLimit | string | `"300Mi"` |  |
| manager.memoryRequest | string | `"200Mi"` |  |
| manager.priorityClassName | string | `nil` |  |
| managerCreateResources | bool | `false` |  |
| managerUserID | int | `8888` |  |
| metrics.port | int | `8080` |  |
| namespaceOverride | string | `""` |  |
| predictiveUnit.defaultEnvSecretRefName | string | `""` |  |
| predictiveUnit.grpcPort | int | `9500` |  |
| predictiveUnit.httpPort | int | `9000` |  |
| predictiveUnit.metricsPortName | string | `"metrics"` |  |
| predictor_servers.HUGGINGFACE_SERVER.protocols.v2.defaultImageVersion | string | `"1.3.0-huggingface"` |  |
| predictor_servers.HUGGINGFACE_SERVER.protocols.v2.image | string | `"seldonio/mlserver"` |  |
| predictor_servers.MLFLOW_SERVER.protocols.seldon.defaultImageVersion | string | `"1.16.0"` |  |
| predictor_servers.MLFLOW_SERVER.protocols.seldon.image | string | `"seldonio/mlflowserver"` |  |
| predictor_servers.MLFLOW_SERVER.protocols.v2.defaultImageVersion | string | `"1.3.0-mlflow"` |  |
| predictor_servers.MLFLOW_SERVER.protocols.v2.image | string | `"seldonio/mlserver"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.seldon.defaultImageVersion | string | `"1.16.0"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.seldon.image | string | `"seldonio/sklearnserver"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.v2.defaultImageVersion | string | `"1.3.0-sklearn"` |  |
| predictor_servers.SKLEARN_SERVER.protocols.v2.image | string | `"seldonio/mlserver"` |  |
| predictor_servers.TEMPO_SERVER.protocols.v2.defaultImageVersion | string | `"1.3.0-slim"` |  |
| predictor_servers.TEMPO_SERVER.protocols.v2.image | string | `"seldonio/mlserver"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.seldon.defaultImageVersion | string | `"1.16.0"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.seldon.image | string | `"seldonio/tfserving-proxy"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.tensorflow.defaultImageVersion | string | `"2.1.0"` |  |
| predictor_servers.TENSORFLOW_SERVER.protocols.tensorflow.image | string | `"tensorflow/serving"` |  |
| predictor_servers.TRITON_SERVER.protocols.v2.defaultImageVersion | string | `"21.08-py3"` |  |
| predictor_servers.TRITON_SERVER.protocols.v2.image | string | `"nvcr.io/nvidia/tritonserver"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.seldon.defaultImageVersion | string | `"1.16.0"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.seldon.image | string | `"seldonio/xgboostserver"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.v2.defaultImageVersion | string | `"1.3.0-xgboost"` |  |
| predictor_servers.XGBOOST_SERVER.protocols.v2.image | string | `"seldonio/mlserver"` |  |
| rbac.configmap.create | bool | `true` |  |
| rbac.create | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `"seldon-manager"` |  |
| singleNamespace | bool | `false` |  |
| storageInitializer.cpuLimit | string | `"1"` |  |
| storageInitializer.cpuRequest | string | `"100m"` |  |
| storageInitializer.image | string | `"seldonio/rclone-storage-initializer:1.16.0"` |  |
| storageInitializer.memoryLimit | string | `"1Gi"` |  |
| storageInitializer.memoryRequest | string | `"100Mi"` |  |
| usageMetrics.enabled | bool | `false` |  |
| webhook.port | int | `4443` |  |
