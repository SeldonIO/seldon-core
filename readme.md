# Documentation

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
 - Expose machine learning models via REST and gRPC automatically when deployed.
 - Allow complex runtime inference graphs to be deployed as microservices. These graphs will be composed of:
   - Models - runtime inference executable for machine learning models
   - Routers - route API requests to sub-graphs. Examples: AB Tests, Multi-Armed Bandits.
   - Combiners - combine the responses from sub-graphs. Examples: ensembles of models
   - Transformers - transform request or responses. Example: transform feature requests.
 - Handle full lifecycle management of the deploy model
    - Updating the runtime graph with no downtime
    - Scaling
    - Monitoring
    - Security

## Quick Start

 - [Quick Start using Minikube](./docs/getting_started/minikube.md)
 - [Jupyter Notebook showing deployment of prebuilt model](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_minikube.ipynb)

## Deployment Guide

 - Wrap your runtime prediction model.
 - Define your runtime inference graph in a seldon deployment custom resource.
 - Deploy.

## Reference

 - [Prediction API](./docs/reference/prediction.md)
 - [Seldon Deployment Custom Resource](./docs/reference/seldon-deployment.md)
