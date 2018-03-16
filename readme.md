# Seldon Core

| Branch      | Status |
|-------------|-------|
| master      | [![Build Status](https://travis-ci.org/SeldonIO/seldon-core.svg?branch=master)](https://travis-ci.org/SeldonIO/seldon-core) |
| release-0.1 | [![Build Status](https://travis-ci.org/SeldonIO/seldon-core.svg?branch=release-0.1)](https://travis-ci.org/SeldonIO/seldon-core) |

Seldon Core is an open source platform for deploying machine learning models on Kubernetes.

- [Goals](#goals)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Install](#install)
- [Deployment guide](#deployment-guide)
- [Reference](#reference)

## Goals

Machine learning deployment has many [challenges](./docs/challenges.md). Seldon Core intends to help with these challenges. Its high level goals are:


 - Allow data scientists to create models using any machine learning toolkit or programming language. We plan to initially cover the tools/languages below:
   - Python based models including
     - Tensorflow models
     - Sklearn models
   - Spark models
   - H2O models
   - R models
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

## Prerequisites

  A [Kubernetes](https://kubernetes.io/) Cluster.  
  Kubernetes can be deployed into many environments, both in cloud and on-premise.

## Quick Start

 - [Quick Start tutorial using Minikube](./docs/getting_started/minikube.md)
 - Jupyter notebooks showing worked examples:
    * Minikube:
         *  [Jupyter Notebook showing deployment of prebuilt model using Minikube](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_minikube.ipynb)
         * [Jupyter notebook to create seldon-core with ksonnet and expose APIs using Ambassador on Minikube.](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/ksonnet_ambassador_minikube.ipynb)
    * GCP:
         * [Jupyter Notebook showing deployment of prebuilt model using GCP cluster](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_gcp.ipynb)
         * [Jupyter notebook to create seldon-core with ksonnet and expose APIs using Ambassador on GCP.](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/ksonnet_ambassador_gcp.ipynb)

## Advanced Tutorials

 * [Advanced graphs](https://github.com/cliveseldon/seldon-core/blob/master/notebooks/advanced_graphs.ipynb) showing the various types of runtime prediction graphs that can be built.

### Example Components
Seldon-core allows various types of components to be built and plugged into the runtime prediction graph. These include [routers, transformers and combiners](docs/reference/internal-api.md). Some components that are available as part of the project are:

 * Models : example that illustrate simple machine learning models to help you build your own integrations
   * [Tensorflow MNIST Classifier](./examples/models/deep_mnist/README.md)

 * Multi-Armed bandits
    * [Epsilon-greedy multi-armed bandits for real time optimization of models](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/epsilon_greedy_gcp.ipynb)
    * Contextual bandits. (coming soon).
 * Outlier Detection
    * [Mahalanobis distance outlier detection](https://github.com/SeldonIO/seldon-core/blob/master/examples/transformers/outlier_mahalanobis/outlier_documentation.ipynb). Example usage can be found in the [Advanced graphs notebook](https://github.com/cliveseldon/seldon-core/blob/master/notebooks/advanced_graphs.ipynb)
 * Model explainers (coming soon).


## Integrations

 * seldon-core can be installed as part of the [kubeflow](https://github.com/kubeflow/kubeflow) project. A detailed [end-to-end example](https://github.com/kubeflow/example-seldon) provides a complete workflow for training various models and deploying them using seldon-core.

## Install

Official releases can be installed via helm from the repository https://storage.googleapis.com/seldon-charts.

To install seldon-core:

```
helm init
helm install seldon-core-crd --name seldon-core-crd --repo https://storage.googleapis.com/seldon-charts
helm install seldon-core --name seldon-core --repo https://storage.googleapis.com/seldon-charts
```

To install the optional analytics components including Prometheus and Grafana with a built-in dashboard for monitoring the running ML deployments run:

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

## Testing

 - [Benchmarking seldon-core](docs/benchmarking.md)

## Community

 * [Slack Channel](https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLWQzMGFkNmRjN2UxZmFmMWJmNWIzMTM5Y2UxNGY1ODE5ZmI2NDdkMmNiMmUxYjZhZGYxOTllMDQwM2NkNDQ1MGI)

## Developer

 - [CHANGELOG](CHANGELOG.md)
 - [Developer Guide](./docs/developer/readme.md)
