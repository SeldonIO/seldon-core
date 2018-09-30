# Seldon Core ![API](./docs/seldon.png)

| Branch      | Status |
|-------------|-------|
| master      | [![Build Status](https://travis-ci.org/SeldonIO/seldon-core.svg?branch=master)](https://travis-ci.org/SeldonIO/seldon-core) |
| release-0.2 | [![Build Status](https://travis-ci.org/SeldonIO/seldon-core.svg?branch=release-0.2)](https://travis-ci.org/SeldonIO/seldon-core) |
| release-0.1 | [![Build Status](https://travis-ci.org/SeldonIO/seldon-core.svg?branch=release-0.1)](https://travis-ci.org/SeldonIO/seldon-core) |


Seldon Core is an open source platform for deploying machine learning models on Kubernetes.

- [Goals](#goals)
- [Quick Start](#quick-start)
- [Example Components](#example-components)
- [Integrations](#integrations)
- [Install](#install)
- [Deployment guide](#deployment-guide)
- [Reference](#reference)
- [Article/Blogs/Videos](#articlesblogsvideos)
- [Community](#community)
- [Developer](#developer)
- [Latest Seldon Images](#latest-seldon-images)
- [Usage Reporting](#usage-reporting)

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

## Important: V1Alpha2 Update

 **We have updated our core API to v1alpha2 which has a breaking change from v1alpha1 in the SeldonDeployments CRD**

[Read details of how to update your kubernetes SeldonDeployment resources](./docs/v1alpha2_update.md).

 * **0.2** releases will now respect the v1alpha2 API.
 * **0.1** releases respect the v1alpha1 API and will not be worked on further.

It is possible to deploy Seldon with two operators that can handle both v1alpha1 resources and v1alpha2 resources though this is not part of our standard deployment docs. If you need this please get in touch.

## Quick Start

Read the [overview to using seldon-core](./docs/getting_started/readme.md).

 - Jupyter notebooks showing worked examples:
    * Minikube:
         * [Jupyter Notebook showing deployment of prebuilt model using Minikube and Helm](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_minikube_rbac.ipynb)
         * [Jupyter Notebook showing deployment of prebuilt model using Minikube, Helm and Ambassador reverse proxy](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/helm_minikube_ambassador.ipynb)	 
         * [Jupyter notebook to create seldon-core with ksonnet and expose APIs using Ambassador on Minikube with RBAC.](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/ksonnet_ambassador_minikube.ipynb)
    * GCP:
         * [Jupyter Notebook showing deployment of prebuilt model using GCP cluster](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_gcp.ipynb)
         * [Jupyter notebook to create seldon-core with ksonnet and expose APIs using Ambassador on GCP.](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/ksonnet_ambassador_gcp.ipynb)
    * Azure
         * [Jupyter Notebook showing deployment of prebuilt model using Azure cluster](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_azure.ipynb)


### Example Components
Seldon-core allows various types of components to be built and plugged into the runtime prediction graph. These include [models, routers, transformers and combiners](docs/reference/internal-api.md). Some example components that are available as part of the project are:

 * **Models** : example that illustrate simple machine learning models to help you build your own integrations
   * Python
      * [Tensorflow MNIST Classifier](./examples/models/deep_mnist/deep_mnist.ipynb)
      * [Keras MNIST Classifier](./examples/models/keras_mnist/keras_mnist.ipynb)
      * [Scikit-learn MNIST Classifier](./examples/models/sk_mnist/skmnist.ipynb)
      * [Scikit-learn Iris Classifier](./examples/models/sklearn_iris/sklearn_iris.ipynb)
   * R
      * [R MNIST Classifier](./examples/models/r_mnist/r_mnist.ipynb)
      * [R Iris Classifier](./examples/models/r_iris/r_iris.ipynb)
   * Java
      * [H2O Classifier](./examples/models/h2o_mojo/h2o_model.ipynb)
   * NodeJS
      * [Tensorflow MNIST Classifier](./examples/models/nodejs_mnist/nodejs_mnist.ipynb)
   * ONNX
      * [ResNet ONNX Classifier using Intel nGraph](./examples/models/onnx_resnet50/onnx_resnet50.ipynb)
   * PMML
      * [PySpark MNIST Classifier](https://github.com/SeldonIO/JPMML-utils/blob/master/examples/pyspark_pmml/mnist.ipynb)

 * **routers**
   * [Epsilon-greedy multi-armed bandits for real time optimization of models](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/epsilon_greedy_gcp.ipynb)
 * **transformers**
    * [Mahalanobis distance outlier detection](https://github.com/SeldonIO/seldon-core/blob/master/examples/transformers/outlier_mahalanobis/outlier_documentation.ipynb). Example usage can be found in the [Advanced graphs notebook](https://github.com/cliveseldon/seldon-core/blob/master/notebooks/advanced_graphs.ipynb)

## Integrations

 * [kubeflow](https://github.com/kubeflow/kubeflow)
    * Seldon-core can be [installed as part of the kubeflow project](https://www.kubeflow.org/docs/guides/components/seldon/#seldon-serving). A detailed [end-to-end example](https://github.com/kubeflow/example-seldon) provides a complete workflow for training various models and deploying them using seldon-core.
 * [IBM's Fabric for Deep Learning](https://github.com/IBM/FfDL)
    * Seldon-core can be used to [serve deep learning models trained using FfDL](https://github.com/IBM/FfDL/blob/master/community/FfDL-Seldon/README.md).
       * [Train and deploy a Tensorflow MNIST classififer using FfDL and Seldon.](https://github.com/IBM/FfDL/blob/master/community/FfDL-Seldon/tf-model/README.md)
       * [Train and deploy a PyTorch MNIST classififer using FfDL and Seldon.](https://github.com/IBM/FfDL/blob/master/community/FfDL-Seldon/pytorch-model/README.md)
 * [Istio and Seldon](./docs/istio.md)
   * [Canary deployemts using Istio and Seldon.](examples/istio/canary_update/canary.ipynb).
 * [NVIDIA TensorRT and DL Inference Server](./integrations/nvidia-inference-server/README.md)
 * [Tensorflow Serving](./integrations/tfserving/README.md)	

## Install

Follow the [install guide](docs/install.md) for details on ways to install seldon onto your Kubernetes cluster.

## Deployment Guide

![API](./docs/deploy.png)

Three steps:

 1. [Wrap your runtime prediction model](./docs/wrappers/readme.md).
    * We provide easy to use wrappers for [python](./docs/wrappers/python.md), [R](./docs/wrappers/r.md) and [Java](./docs/wrappers/java.md)
 1. [Define your runtime inference graph in a seldon deployment custom resource](./docs/inference-graph.md).
 1. [Deploy the graph](./docs/deploying.md).

## Advanced Tutorials

 * [Advanced graphs](https://github.com/seldonio/seldon-core/blob/master/notebooks/advanced_graphs.ipynb) showing the various types of runtime prediction graphs that can be built.
 * [Handling large gRPC messages](./notebooks/max_grpc_msg_size.ipynb). Showing how you can add annotations to increase the gRPC max message size.
 * [Handling REST timeouts](./notebooks/timeouts.ipynb). Showing how you can add annotations to set the REST (and gRPC) timeouts.

## Reference

 - [Prediction API](./docs/reference/prediction.md)
 - [Seldon Deployment Custom Resource](./docs/reference/seldon-deployment.md)
 - [Analytics](./docs/analytics.md)
 
## Articles/Blogs/Videos

 - [Kubecon Europe 2018 - Serving Machine Learning Models at Scale with Kubeflow and Seldon](https://www.youtube.com/watch?v=pDlapGtecbY)
 - [AI on Kubernetes - O'Reilly Tutorial](https://github.com/dwhitena/oreilly-ai-k8s-tutorial)
 - [Scalable Data Science - The State of DevOps/MLOps in 2018](https://axsauze.github.io/scalable-data-science/#/)
 - [Istio Weekly Community Meeting - Seldon-core with Istio](https://www.youtube.com/watch?v=ydculT4e7FQ&feature=youtu.be&t=7m48s)
 - [Openshift Commons ML SIG - Openshift S2I Helping ML Deployment with Seldon-Core](https://www.youtube.com/watch?v=1uZPBcfYxlM)
 - [Overview of Openshift source-to-image use in Seldon-Core](./docs/articles/openshift_s2i.md)
 - [IBM Framework for Deep Learning and Seldon-Core](https://developer.ibm.com/code/2018/06/12/serve-it-hot-deploy-your-ffdl-trained-models-using-seldon/)
 - [CartPole game by Reinforcement Learning, a journey from training to inference ](https://github.com/hypnosapos/cartpole-rl-remote/)

### Release Highlights

 * [0.2.3 Release Highlights](docs/articles/release-0.2.3.md)

## Testing

 - [Benchmarking seldon-core](docs/benchmarking.md)

## Configuration

 - [Annotation based configuration](./docs/annotations.md).
 - [Notes for running in production](./docs/production.md).
 - [Helm configuration](./docs/helm.md)
 - [ksonnet configuration](./docs/ksonnet.md)

## Community

 * [Slack Channel](https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLWQzMGFkNmRjN2UxZmFmMWJmNWIzMTM5Y2UxNGY1ODE5ZmI2NDdkMmNiMmUxYjZhZGYxOTllMDQwM2NkNDQ1MGI)

## Developer

 - [CHANGELOG](CHANGELOG.md)
 - [Developer Guide](./docs/developer/readme.md)

## Latest Seldon Images

| Description | Image URL | Stable Version | Development | 
|-------------|-----------|----------------|-----|
| Seldon Operator | [seldonio/cluster-manager](https://hub.docker.com/r/seldonio/cluster-manager/tags/) | 0.2.3 | 0.2.4-SNAPSHOT |
| Seldon Service Orchestrator | [seldonio/engine](https://hub.docker.com/r/seldonio/engine/tags/) | 0.2.3 | 0.2.4-SNAPSHOT |
| Seldon API Gateway | [seldonio/apife](https://hub.docker.com/r/seldonio/apife/tags/) | 0.2.3 | 0.2.4-SNAPSHOT |
| [Seldon Python 3 Wrapper for S2I](docs/wrappers/python.md) | [seldonio/seldon-core-s2i-python3](https://hub.docker.com/r/seldonio/seldon-core-s2i-python3/tags/) | 0.2 |  |
| [Seldon Python 2 Wrapper for S2I](docs/wrappers/python.md) | [seldonio/seldon-core-s2i-python2](https://hub.docker.com/r/seldonio/seldon-core-s2i-python2/tags/) | 0.2 |  |
| [Seldon Python ONNX Wrapper for S2I](docs/wrappers/python.md) | [seldonio/seldon-core-s2i-python3-ngraph-onnx](https://hub.docker.com/r/seldonio/seldon-core-s2i-python3-ngraph-onnx/tags/) | 0.1  |   |
| [Seldon Core Python Wrapper](docs/wrappers/python-docker.md) | [seldonio/core-python-wrapper](https://hub.docker.com/r/seldonio/core-python-wrapper/tags/) | 0.7 | |
| [Seldon Java Build Wrapper for S2I](docs/wrappers/java.md) | [seldonio/seldon-core-s2i-java-build](https://hub.docker.com/r/seldonio/seldon-core-s2i-java-build/tags/) | 0.1 | |
| [Seldon Java Runtime Wrapper for S2I](docs/wrappers/java.md) | [seldonio/seldon-core-s2i-java-runtime](https://hub.docker.com/r/seldonio/seldon-core-s2i-java-runtime/tags/) | 0.1 | | 
| [Seldon R Wrapper for S2I](docs/wrappers/r.md) | [seldonio/seldon-core-s2i-r](https://hub.docker.com/r/seldonio/seldon-core-s2i-r/tags/) | 0.1 | |
| [Seldon NodeJS Wrapper for S2I](docs/wrappers/nodejs.md) | [seldonio/seldon-core-s2i-nodejs](https://hub.docker.com/r/seldonio/seldon-core-s2i-nodejs/tags/) | 0.1 | |
| [Seldon Tensorflow Serving proxy](integrations/tfserving/README.md) | [seldonio/tfserving-proxy](https://hub.docker.com/r/seldonio/tfserving-proxy/tags/) | 0.1 | 
| [Seldon NVIDIA inference server proxy](integrations/nvidia-inference-server/README.md) | [seldonio/nvidia-inference-server-proxy](https://hub.docker.com/r/seldonio/nvidia-inference-server-proxy/tags/) | 0.1 | 
#### Java Packages

| Description | Package | Version |
|-------------|---------|---------|
| [Seldon Core Wrapper](https://github.com/SeldonIO/seldon-java-wrapper) | [seldon-core-wrapper](https://mvnrepository.com/artifact/io.seldon.wrapper/seldon-core-wrapper) | 0.1.2 |
| [Seldon Core JPMML](https://github.com/SeldonIO/JPMML-utils) | [seldon-core-jpmml](https://mvnrepository.com/artifact/io.seldon.wrapper/seldon-core-jpmml) | 0.0.1 |

## Usage Reporting

Tools that help the development of Seldon Core from anonymous usage.
* [Usage Reporting with Spartakus](docs/usage-reporting.md)


