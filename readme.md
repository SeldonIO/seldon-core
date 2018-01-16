# Seldon Core

Seldon Core is an open source framework for deploying machine learning models on Kubernetes.

- [Goals](#goals)
- [Quick Start](#quick-start)
- [Deployment guide](#deployment-guide)
- [Reference](#reference)

## Goals

Seldon Core goals:

 - Allow data scientists to create models using any machine learning toolkit or programming language. We plan to initially cover the tools/languages below:
   - Python based models including
     - Tensorflow models
     - Sklearn models
   - Spark Models
   - H2O Models
 - Expose machine learning models via REST and gRPC automatically when deployed for easy integration into business apps that need predictions.
 - Allow complex runtime inference graphs to be deployed as microservices. These graphs can be composed of:
   - Models - runtime inference executable for machine learning models
   - Routers - route API requests to sub-graphs. Examples: AB Tests, Multi-Armed Bandits.
   - Combiners - combine the responses from sub-graphs. Examples: ensembles of models
   - Transformers - transform request or responses. Example: transform feature requests.
 - Handle full lifecycle management of the deployed model:
    - Updating the runtime graph with no downtime
    - Scaling
    - Monitoring
    - Security

## Quick Start

 - [Quick Start using Minikube](./docs/getting_started/minikube.md)
 - [Jupyter Notebook showing deployment of prebuilt model using Minikube](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_minikube.ipynb)
 - [Jupyter Notebook showing deployment of prebuilt model using GCP cluster](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_gcp.ipynb)

## Installation

Official releases can be installed via helm from the repository https://storage.googleapis.com/seldon-charts.
For example:

```
helm install seldon-core --name seldon-core \
    --set grafana_prom_admin_password=password \
    --set persistence.enabled=false \
    --repo https://storage.googleapis.com/seldon-charts
```

## Deployment Guide

![API](./docs/deploy.png)

Three steps:

 1. [Wrap your runtime prediction model](./docs/wrappers/readme.md).
 1. [Define your runtime inference graph in a seldon deployment custom resource](./docs/crd/readme.md).
 1. [Deploy the graph](./docs/deploying.md).

## Reference

 - [Prediction API](./docs/reference/prediction.md)
 - [Seldon Deployment Custom Resource](./docs/reference/seldon-deployment.md)


## Developer

 - [CHANGELOG](CHANGELOG.md)
 - [Developer Guide](./docs/developer/readme.md)
