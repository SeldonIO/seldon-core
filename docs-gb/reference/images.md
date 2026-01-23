# Latest Seldon Images


## Core images

| Description | Image URL | Version |
|-------------|-----------|--------------|
| [Seldon Operator](../workflow/install.md) | [seldonio/seldon-core-operator](https://hub.docker.com/r/seldonio/seldon-core-operator/tags/) | 1.19.0 |
| [Seldon Service Orchestrator (Go)](../graph/svcorch.md)| [seldonio/seldon-core-executor](https://hub.docker.com/r/seldonio/seldon-core-executor/tags) | 1.19.0 |

## Pre-packaged servers


| Description | Image URL | Version |
|-------------|-----------|---------|
| [MLFlow Server](../servers/mlflow.md) | [seldonio/mlflowserver](https://hub.docker.com/r/seldonio/mlflowserver/tags/) | 1.19.0 |
| [SKLearn Server](../servers/sklearn.md) | [seldonio/sklearnserver](https://hub.docker.com/r/seldonio/sklearnserver/tags/) | 1.19.0 |
| [XGBoost Server](../servers/xgboost.md) | [seldonio/xgboostserver](https://hub.docker.com/r/seldonio/xgboostserver/tags/) | 1.19.0 |

## Language wrappers

| Description | Image URL | Version |
|-------------|-----------|----------------|
| [Seldon Python 3 (3.12) Wrapper for S2I](../wrappers/python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python3](https://hub.docker.com/r/seldonio/seldon-core-s2i-python3/tags/) | 1.19.0 |
| [Seldon Python 3.12 Wrapper for S2I](../wrappers/python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python312](https://hub.docker.com/r/seldonio/seldon-core-s2i-python312/tags/) | 1.19.0 |

## Server proxies

| Description | Image URL | Version |
|-------------|-----------|----------------|
| [SageMaker proxy](https://github.com/SeldonIO/seldon-core/tree/master/integrations/sagemaker) | [seldonio/sagemaker-proxy](https://hub.docker.com/r/seldonio/sagemaker-proxy/tags/) | 0.1 |
| [Tensorflow Serving proxy](../servers/tensorflow.md) | [seldonio/tfserving-proxy](https://hub.docker.com/r/seldonio/tfserving-proxy/tags/) | 1.19.0 |


## Python modules

| Description | Python Version | Version |
|-------------|----------------|---------|
| [seldon-core](https://pypi.org/project/seldon-core/) | >3.4,<3.13 | 1.19.0 |


## Other images 

{% hint style="warning" %}
The following images are examples only and are not production-ready.
{% endhint %}

### Language wrappers (examples only)

| Description | Image URL  |
|-------------|-----------|
| [Seldon Python ONNX Wrapper for S2I](../wrappers/python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python3-ngraph-onnx](https://hub.docker.com/r/seldonio/seldon-core-s2i-python3-ngraph-onnx/tags/) |
| [Seldon Java Build Wrapper for S2I](../wrappers/java.md) | [seldonio/seldon-core-s2i-java-build](https://hub.docker.com/r/seldonio/seldon-core-s2i-java-build/tags/) |
| [Seldon Java Runtime Wrapper for S2I](../wrappers/java.md) | [seldonio/seldon-core-s2i-java-runtime](https://hub.docker.com/r/seldonio/seldon-core-s2i-java-runtime/tags/) |
| [Seldon R Wrapper for S2I](../wrappers/R.md) | [seldonio/seldon-core-s2i-r](https://hub.docker.com/r/seldonio/seldon-core-s2i-r/tags/) |
| [Seldon NodeJS Wrapper for S2I](../wrappers/nodejs.md) | [seldonio/seldon-core-s2i-nodejs](https://hub.docker.com/r/seldonio/seldon-core-s2i-nodejs/tags/) |

### Java packages (examples only)

You can find these packages in the Maven repository.

| Description | Package | Version |
|-------------|---------|---------|
| [Seldon Core Wrapper](https://github.com/SeldonIO/seldon-java-wrapper) | seldon-core-wrapper | 0.1.5 |
| [Seldon Core JPMML](https://github.com/SeldonIO/JPMML-utils) | seldon-core-jpmml | 0.0.1 |
