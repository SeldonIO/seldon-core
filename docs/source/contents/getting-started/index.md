# Getting Started


 * [Install locally with Docker Compose](./docker-installation/index.md)
 * [Install onto a Kubernetes cluster](./kubernetes-installation/index.md)


## Core Concepts

There are three core resources you will use:

 * Models - for deploying single machine learning models, custom transformation logic, drift detectors and outliers detectors.
 * Pipelines - for connecting together flows of data between models
 * Experiments - for testing new versions of models
 * Explainers - for explaining model or pipeline  predictions

By default the standard installation will deploy MLServer and Triton inference servers which provide support for a wide range of machine learning model artifacts including Tensorflow models, PyTorch models, SKlearn models, XGBoost models, ONNX models, TensorRT models, custom python models and many more. For advanced use, the creation of new inference servers is manged by two resources:

 * Servers - for deploying sets of replicas of core inference servers (MLServer or Triton).
 * ServerConfigs - for defining server configurations

## API for Inference

Once deployed models can be called using the Seldon V2 inference protocol. This protocol created by Seldon, NVIDIA and the KServe projects is supported by MLServer and Triton inference servers and allows REST and gRPC calls to your model.

Your model is exposed via our internal Envoy gateway. If you wish to expose your models in Kubernetes outside the cluster you are free to use any Service Mesh or Ingress technology. Various examples are provided for service mesh integration.

## Inference Metrics

Metrics are exposed for scrapping by Prometheus. For Kubernetes we provide example instructions for using kube-prometheus.

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