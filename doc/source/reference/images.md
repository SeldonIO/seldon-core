# Latest Seldon Images


## Core images

| Description | Image URL | Stable Version | Development |
|-------------|-----------|----------------|-------------|
| [Seldon Operator](../workflow/install.md) | [seldonio/seldon-core-operator](https://hub.docker.com/r/seldonio/seldon-core-operator/tags/) | 1.10.0 | 1.11.0-dev |
| [Seldon Service Orchestrator (Go)](../graph/svcorch.md)| [seldonio/seldon-core-executor](https://hub.docker.com/r/seldonio/executor/tags/) | 1.10.0 | 1.11.0-dev |

## Pre-packaged servers


| Description | Image URL | Version |
|-------------|-----------|---------|
| [MLFlow Server REST](../servers/mlflow.md) | [seldonio/mlflowserver_rest](https://hub.docker.com/r/seldonio/mlflowserver_rest/tags/) | 1.10.0 |
| [MLFlow Server GRPC](../servers/mlflow.md) | [seldonio/mlflowserver_grpc](https://hub.docker.com/r/seldonio/mlflowserver_grpc/tags/) | 1.10.0 |
| [SKLearn Server REST](../servers/sklearn.md) | [seldonio/sklearnserver_rest](https://hub.docker.com/r/seldonio/sklearnserver_rest/tags/) | 1.10.0 |
| [SKLearn Server GRPC](../servers/sklearn.md) | [seldonio/sklearnserver_grpc](https://hub.docker.com/r/seldonio/sklearnserver_grpc/tags/) | 1.10.0 |
| [XGBoost Server REST](../servers/xgboost.md) | [seldonio/xgboostserver_rest](https://hub.docker.com/r/seldonio/xgboostserver_rest/tags/) | 1.10.0 |
| [XGBoost Server GRPC](../servers/xgboost.md) | [seldonio/xgboostserver_grpc](https://hub.docker.com/r/seldonio/xgboostserver_grpc/tags/) | 1.10.0 |

## Language wrappers

| Description | Image URL | Stable Version | Development |
|-------------|-----------|----------------|-------------|
| [Seldon Python 3 (3.6) Wrapper for S2I](../python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python3](https://hub.docker.com/r/seldonio/seldon-core-s2i-python3/tags/) | 1.10.0 | 1.11.0-dev |
| [Seldon Python 3.6 Wrapper for S2I](../python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python36](https://hub.docker.com/r/seldonio/seldon-core-s2i-python36/tags/) | 1.10.0 | 1.11.0-dev |
| [Seldon Python 3.7 Wrapper for S2I](../python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python37](https://hub.docker.com/r/seldonio/seldon-core-s2i-python37/tags/) | 1.10.0 | 1.11.0-dev |
| [Seldon Python 3.6 GPU Wrapper for S2I](../python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python36-gpu](https://hub.docker.com/r/seldonio/seldon-core-s2i-python36-gpu/tags/) | 1.10.0 | 1.11.0-dev |
| [Seldon Python 3.7 GPU Wrapper for S2I](../python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python37-gpu](https://hub.docker.com/r/seldonio/seldon-core-s2i-python37-gpu/tags/) | 1.10.0 | 1.11.0-dev |

## Server proxies

| Description | Image URL | Stable Version |
|-------------|-----------|----------------|
| [NVIDIA inference server proxy](integration_nvidia_link.rst) | [seldonio/nvidia-inference-server-proxy](https://hub.docker.com/r/seldonio/nvidia-inference-server-proxy/tags/) | 0.1 |
| [SageMaker proxy](https://github.com/SeldonIO/seldon-core/tree/master/integrations/sagemaker) | [seldonio/sagemaker-proxy](https://hub.docker.com/r/seldonio/sagemaker-proxy/tags/) | 0.1 |
| [Tensorflow Serving REST proxy](../servers/tensorflow.md) | [seldonio/tfserving-proxy_rest](https://hub.docker.com/r/seldonio/tfserving-proxy_rest/tags/) | 0.7 |
| [Tensorflow Serving GRPC proxy](../servers/tensorflow.md) | [seldonio/tfserving-proxy_grpc](https://hub.docker.com/r/seldonio/tfserving-proxy_grpc/tags/) | 0.7 |


## Python modules

| Description | Python Version | Version |
|-------------|----------------|---------|
| [seldon-core](https://pypi.org/project/seldon-core/) | >3.4,<3.7 | 1.10.0 |
| [seldon-core](https://pypi.org/project/seldon-core/) | 2,>=3,<3.7 | 0.2.6 (deprecated) |


## Incubating

### Language wrappers

| Description | Image URL | Stable Version | Development |
|-------------|-----------|----------------|-------------|
| [Seldon Python ONNX Wrapper for S2I](../python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python3-ngraph-onnx](https://hub.docker.com/r/seldonio/seldon-core-s2i-python3-ngraph-onnx/tags/) | 0.3  |   |
| [Seldon Java Build Wrapper for S2I](../java/README.md) | [seldonio/seldon-core-s2i-java-build](https://hub.docker.com/r/seldonio/seldon-core-s2i-java-build/tags/) | 0.1 | |
| [Seldon Java Runtime Wrapper for S2I](../java/README.md) | [seldonio/seldon-core-s2i-java-runtime](https://hub.docker.com/r/seldonio/seldon-core-s2i-java-runtime/tags/) | 0.1 | |
| [Seldon R Wrapper for S2I](../R/README.md) | [seldonio/seldon-core-s2i-r](https://hub.docker.com/r/seldonio/seldon-core-s2i-r/tags/) | 0.2 | |
| [Seldon NodeJS Wrapper for S2I](../nodejs/README.md) | [seldonio/seldon-core-s2i-nodejs](https://hub.docker.com/r/seldonio/seldon-core-s2i-nodejs/tags/) | 0.1 | 0.2-SNAPSHOT |


### Java packages

| Description | Package | Version |
|-------------|---------|---------|
| [Seldon Core Wrapper](https://github.com/SeldonIO/seldon-java-wrapper) | [seldon-core-wrapper](https://mvnrepository.com/artifact/io.seldon.wrapper/seldon-core-wrapper) | 0.1.5 |
| [Seldon Core JPMML](https://github.com/SeldonIO/JPMML-utils) | [seldon-core-jpmml](https://mvnrepository.com/artifact/io.seldon.wrapper/seldon-core-jpmml) | 0.0.1 |



## Deprecated

### Language wrappers

| Description | Image URL | Stable Version | Development |
|-------------|-----------|----------------|-------------|
| [Seldon Python 2 Wrapper for S2I](../python/python_wrapping_s2i.md) | [seldonio/seldon-core-s2i-python2](https://hub.docker.com/r/seldonio/seldon-core-s2i-python2/tags/) | 0.5.1 | deprecated |
