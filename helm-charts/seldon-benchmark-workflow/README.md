# seldon-benchmark-workflow

![Version: 0.1](https://img.shields.io/static/v1?label=Version&message=0.1&color=informational&style=flat-square)

Seldon Benchmark Workflow

## Usage

To use this chart, you will first need to add the `seldonio` Helm repo:

```bash
helm repo add seldonio https://storage.googleapis.com/seldon-charts
helm repo update
```

Once that's done, you should then be able to use the inference graph template as:

```bash
helm template $MY_MODEL_NAME seldonio/seldon-benchmark-workflow --namespace $MODELS_NAMESPACE
```

Note that you can also deploy the inference graph directly to your cluster
using:

```bash
helm install $MY_MODEL_NAME seldonio/seldon-benchmark-workflow --namespace $MODELS_NAMESPACE
```

## Source Code

* <https://github.com/SeldonIO/seldon-core>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| benchmark.concurrency | int | `1` |  |
| benchmark.cpu | int | `4` |  |
| benchmark.data | string | `"{\"data\": {\"ndarray\": [[0,1,2,3]]}}"` |  |
| benchmark.duration | string | `"30s"` |  |
| benchmark.grpcDataOverride | string | `nil` |  |
| benchmark.grpcImage | string | `"seldonio/ghz:v0.95.0"` |  |
| benchmark.host | string | `"istio-ingressgateway.istio-system.svc.cluster.local:80"` |  |
| benchmark.rate | int | `0` |  |
| benchmark.restImage | string | `"peterevans/vegeta:latest-vegeta12.8.4"` |  |
| seldonDeployment.apiType | string | `"rest"` |  |
| seldonDeployment.disableOrchestrator | bool | `false` |  |
| seldonDeployment.enableResources | string | `"false"` |  |
| seldonDeployment.image | string | `nil` |  |
| seldonDeployment.limits.cpu | string | `"50m"` |  |
| seldonDeployment.limits.memory | string | `"1000Mi"` |  |
| seldonDeployment.modelName | string | `"classifier"` |  |
| seldonDeployment.modelUri | string | `nil` |  |
| seldonDeployment.name | string | `"seldon-{{workflow.uid}}"` |  |
| seldonDeployment.protocol | string | `"seldon"` |  |
| seldonDeployment.replicas | int | `2` |  |
| seldonDeployment.requests.cpu | string | `"50m"` |  |
| seldonDeployment.requests.memory | string | `"100Mi"` |  |
| seldonDeployment.server | string | `nil` |  |
| seldonDeployment.serverThreads | int | `1` |  |
| seldonDeployment.serverWorkers | int | `4` |  |
| seldonDeployment.waitTime | int | `5` |  |
| workflow.name | string | `"seldon-benchmark-process"` |  |
| workflow.namespace | string | `"default"` |  |
| workflow.parallelism | int | `1` |  |
| workflow.paramDelimiter | string | `"|"` |  |
| workflow.useNameAsGenerateName | string | `"false"` |  |
