# Getting Started

```{note}
Some dependencies may require that the (virtual) machines on which you deploy, support the SSE4.2 instruction set or x86-64-v2 microarchitecture. If `lscpu | grep sse4_2` does not return anything on your machine, your CPU is not compatible, and you may need to update the (virtual) host's CPU.
```

Seldon Core can be installed either with Docker Compose or with Kubernetes:

 * [Install locally with Docker Compose](./docker-installation/index.md)
 * [Install onto a Kubernetes cluster](./kubernetes-installation/index.md)

Once installed:

  * Try the existing [examples](../examples/index.md).
  * Train and deploy your own [model artifact](../models/inference-artifacts/index.md#saving-model-artifacts).


## Core Concepts

There are three core resources you will use:

 * [Models](../models/index.md) - for deploying single machine learning models, custom transformation logic, drift detectors and outliers detectors.
 * [Pipelines](../pipelines/index.md) - for connecting together flows of data transformations between Models with a synchronous path and multiple asynchronous paths.
 * [Experiments](../experiments/index.md) - for testing new versions of models

By default the standard installation will deploy MLServer and Triton inference servers which provide support for a wide range of machine learning model artifacts including Tensorflow models, PyTorch models, SKlearn models, XGBoost models, ONNX models, TensorRT models, custom python models and many more. For advanced use, the creation of new inference servers is manged by two resources:

 * [Servers](../servers/index.md) - for deploying sets of replicas of core inference servers (MLServer or Triton by default).
 * [ServerConfigs](../kubernetes/resources/serverconfig/index.md) - for defining server configurations including custom servers.

## API for Inference

Once deployed models can be called using the [Seldon V2 inference protocol](../apis/inference/v2.md). This protocol created by Seldon, NVIDIA and the KServe projects is supported by MLServer and Triton inference servers amingst others and allows REST and gRPC calls to your model.

Your model is exposed via our internal Envoy gateway. If you wish to expose your models in Kubernetes outside the cluster you are free to use any Service Mesh or Ingress technology. Various examples are provided for service mesh integration.

## Inference Metrics

Metrics are exposed for scraping by Prometheus. For Kubernetes we provide example instructions for using kube-prometheus.

## Pipeline requirements

Pipelines are built upon Kafka streaming technology.


```{toctree}
:maxdepth: 1
:hidden:

docker-installation/index.md
kubernetes-installation/index.md
configuration/index.md
cli.md
```
