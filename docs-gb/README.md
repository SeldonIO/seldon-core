# Quick Start Guide

A platform to deploy your machine learning models on Kubernetes at massive scale.

## Seldon Core V2 Now Available

[![scv2\_image](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/_static/scv2_banner.png)](https://docs.seldon.ai/seldon-core-2)

[Seldon Core V2](https://docs.seldon.ai/seldon-core-2) **is now available**. If you're new to Seldon Core we recommend you [start here](https://docs.seldon.ai/seldon-core-2/installation/learning-environment). Check out the [docs here](https://docs.seldon.ai/seldon-core-2) and make sure to leave feedback on [our slack community](https://join.slack.com/t/seldondev/shared_invite/zt-vejg6ttd-ksZiQs3O_HOtPQsen_labg) and [submit bugs or feature requests on the repo](https://github.com/SeldonIO/seldon-core/issues/new/choose). The codebase can be found [in this branch](https://github.com/SeldonIO/seldon-core/tree/v2).

Continue reading for info on Seldon Core V1...

[![video\_play\_icon](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/core-play-logo.png)](https://www.youtube.com/watch?v=5Q-03We8aDE)

## Overview

Seldon core converts your ML models (Tensorflow, Pytorch, H2o, etc.) or language wrappers (Python, Java, etc.) into production REST/GRPC microservices.

Seldon handles scaling to thousands of production machine learning models and provides advanced machine learning capabilities out of the box including Advanced Metrics, Request Logging, Explainers, Outlier Detectors, A/B Tests, Canaries and more.

* Read the [Seldon Core Documentation](https://docs.seldon.ai/seldon-core-1)
* Join our [community Slack](https://join.slack.com/t/seldondev/shared_invite/zt-vejg6ttd-ksZiQs3O_HOtPQsen_labg) to ask any questions
* Get started with [Seldon Core Notebook Examples](notebooks/)
* Learn how you can [start contributing](developer/contributing.md)

![](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/seldon-core-high-level.jpg)

### High Level Features

With over 2M installs, Seldon Core is used across organisations to manage large scale deployment of machine learning models, and key benefits include:

* Easy way to containerise ML models using our [pre-packaged inference servers](servers/overview.md), [custom servers](servers/custom.md), or [language wrappers](https://docs.seldon.io/projects/seldon-core/en/1.18/wrappers/language_wrappers.html).
* Out of the box endpoints which can be tested through [Swagger UI](https://docs.seldon.io/projects/seldon-core/en/latest/reference/apis/openapi.html?highlight=swagger), [Seldon Python Client or Curl / GRPCurl](wrappers/python/python_module.md#seldon-core-python-package).
* Cloud agnostic and tested on [AWS EKS, Azure AKS, Google GKE, Alicloud, Digital Ocean and Openshift](notebooks/#cloud-specific-examples).
* Powerful and rich inference graphs made out of [predictors, transformers, routers, combiners, and more](https://docs.seldon.io/projects/seldon-core/en/1.18/examples/graph-metadata.html).
* Metadata provenance to ensure each model can be traced back to its respective [training system, data and metrics](deployments/metadata.md)
* Advanced and customisable metrics with integration [to Prometheus and Grafana](integrations/analytics.md).
* Full auditability through model input-output request [logging integration with Elasticsearch](reference/log_level.md).
* Microservice distributed tracing through [integration to Jaeger](integrations/distributed-tracing.md) for insights on latency across microservice hops.
* Secure, reliable and robust system maintained through a consistent [security & updates policy](../SECURITY.md).

## Getting Started

Deploying your models using Seldon Core is simplified through our pre-packaged inference servers and language wrappers. Below you can see how you can deploy our "hello world Iris" example. You can see more details on these workflows in our [Documentation Quickstart](./#getting-started).

### Install Seldon Core

Quick install using Helm 3 (you can also use Kustomize):

```bash
kubectl create namespace seldon-system

helm install seldon-core seldon-core-operator \
    --repo https://storage.googleapis.com/seldon-charts \
    --set usageMetrics.enabled=true \
    --namespace seldon-system \
    --set istio.enabled=true
    # You can set ambassador instead with --set ambassador.enabled=true
```

### Deploy your model using pre-packaged model servers

We provide optimized model servers for some of the most popular Deep Learning and Machine Learning frameworks that allow you to deploy your trained model binaries/weights without having to containerize or modify them.

You only have to upload your model binaries into your preferred object store, in this case we have a trained scikit-learn iris model in a Google bucket:

```console
gs://seldon-models/v1.19.0-dev/sklearn/iris/model.joblib
```

Create a namespace to run your model in:

```bash
kubectl create namespace seldon
```

We then can deploy this model with Seldon Core to our Kubernetes cluster using the pre-packaged model server for scikit-learn (SKLEARN\_SERVER) by running the `kubectl apply` command below:

```yaml
$ kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: iris-model
  namespace: seldon
spec:
  name: iris
  predictors:
  - graph:
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.19.0-dev/sklearn/iris
      name: classifier
    name: default
    replicas: 1
END
```

#### Send API requests to your deployed model

Every model deployed exposes a standardised User Interface to send requests using our OpenAPI schema.

This can be accessed through the endpoint `http://<ingress_url>/seldon/<namespace>/<model-name>/api/v1.0/doc/` which will allow you to send requests directly through your browser.

![](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/rest-openapi.jpg)

Or alternatively you can send requests programmatically using our [Seldon Python Client](wrappers/python/seldon_client.md) or another Linux CLI:

```console
$ curl -X POST http://<ingress>/seldon/seldon/iris-model/api/v1.0/predictions \
    -H 'Content-Type: application/json' \
    -d '{ "data": { "ndarray": [[1,2,3,4]] } }'

{
   "meta" : {},
   "data" : {
      "names" : [
         "t:0",
         "t:1",
         "t:2"
      ],
      "ndarray" : [
         [
            0.000698519453116284,
            0.00366803903943576,
            0.995633441507448
         ]
      ]
   }
}
```

### Deploy your custom model using language wrappers

For more custom deep learning and machine learning use-cases which have custom dependencies (such as 3rd party libraries, operating system binaries or even external systems), we can use any of the Seldon Core language wrappers.

You only have to write a class wrapper that exposes the logic of your model; for example in Python we can create a file `Model.py`:

```python
import pickle
class Model:
    def __init__(self):
        self._model = pickle.loads( open("model.pickle", "rb") )

    def predict(self, X):
        output = self._model(X)
        return output
```

We can now containerize our class file using the [Seldon Core s2i utils](https://docs.seldon.io/projects/seldon-core/en/1.18/wrappers/s2i.html) to produce the `sklearn_iris` image:

```console
s2i build . seldonio/seldon-core-s2i-python3:0.18 sklearn_iris:0.1
```

And we now deploy it to our Seldon Core Kubernetes Cluster:

```yaml
$ kubectl apply -f - << END
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: iris-model
  namespace: model-namespace
spec:
  name: iris
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - name: classifier
          image: sklearn_iris:0.1
    graph:
      name: classifier
    name: default
    replicas: 1
END
```

#### Send API requests to your deployed model

Every model deployed exposes a standardised User Interface to send requests using our OpenAPI schema.

This can be accessed through the endpoint `http://<ingress_url>/seldon/<namespace>/<model-name>/api/v1.0/doc/` which will allow you to send requests directly through your browser.

![](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/rest-openapi.jpg)

Or alternatively you can send requests programmatically using our [Seldon Python Client](wrappers/python/seldon_client.md) or another Linux CLI:

```console
$ curl -X POST http://<ingress>/seldon/model-namespace/iris-model/api/v1.0/predictions \
    -H 'Content-Type: application/json' \
    -d '{ "data": { "ndarray": [1,2,3,4] } }' | json_pp

{
   "meta" : {},
   "data" : {
      "names" : [
         "t:0",
         "t:1",
         "t:2"
      ],
      "ndarray" : [
         [
            0.000698519453116284,
            0.00366803903943576,
            0.995633441507448
         ]
      ]
   }
}
```

### Dive into the Advanced Production ML Integrations

Any model that is deployed and orchestrated with Seldon Core provides out of the box machine learning insights for monitoring, managing, scaling and debugging.

Below are some of the core components together with link to the logs that provide further insights on how to set them up.

| <p><a href="https://docs.seldon.ai/seldon-core-1/configuration/integrations/analytics"><br>Standard and custom metrics with prometheus<br><br><img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/analytics/dashboard.png" alt=""></a></p>  | <p><a href="https://docs.seldon.ai/seldon-core-1/configuration/integrations/logging"><br>Full audit trails with ELK request logging<br><br><img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/kibana-custom-search.png" alt=""></a></p>                  |
| -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| <p><a href="https://docs.seldon.ai/seldon-core-1/configuration/integrations/explainers"><br>Explainers for Machine Learning Interpretability<br><br><img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/anchors.jpg" alt=""></a></p> | <p><a href="https://docs.seldon.ai/seldon-core-1/configuration/integrations/outlier_detection"><br>Outlier and Adversarial Detectors for Monitoring<br><br><img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/adversarial-attack.png" alt=""></a></p>    |
| <p><a href="https://docs.seldon.ai/seldon-core-1/configuration/integrations/cicd-mlops"><br>CI/CD for MLOps at Massive Scale<br><br><img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/cicd-seldon.jpg" alt=""></a></p>             | <p><a href="https://docs.seldon.ai/seldon-core-1/configuration/integrations/distributed-tracing"><br>Distributed tracing for performance monitoring<br><br><img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/graph/jaeger-ui-rest-example.png" alt=""></a></p> |

## Where to go from here

### Getting Started

* [Quickstart Guide](./#getting-started)
* [Overview of Components](overview.md)
* [Install Seldon Core on Kubernetes](../install/installation.md)
* [Join the Community](developer/community.md)

### Seldon Core Deep Dive

* [Detailed Installation Parameters](https://docs.seldon.io/projects/seldon-core/en/latest/reference/helm.html)
* [Pre-packaged Inference Servers](servers/overview.md)
* [Language Wrappers for Custom Models](https://docs.seldon.io/projects/seldon-core/en/1.18/wrappers/language_wrappers.html)
* [Create your Inference Graph](routing/inference-graph.md)
* [Deploy your Model](deployments/deploying.md)
* [Testing your Model Endpoints](deployments/serving.md)
* [Troubleshooting guide](deployments/troubleshooting.md)
* [Usage reporting](install/usage-reporting.md)
* [Upgrading](upgrading.md)
* [Changelog](../CHANGELOG.md)

### Pre-Packaged Inference Servers

* [MLflow Server](servers/mlflow.md)
* [SKLearn server](servers/sklearn.md)
* [Tensorflow Serving](servers/tensorflow.md)
* [XGBoost server](servers/xgboost.md)

### Language Wrappers (Production)

* [Python Language Wrapper \[Production\]](https://docs.seldon.io/projects/seldon-core/en/latest/python/index.html)

### Language Wrappers (Incubating)

* [Java Language Wrapper \[Incubating\]](wrappers/java.md)
* [R Language Wrapper \[ALPHA\]](wrappers/r.md)
* [NodeJS Language Wrapper \[ALPHA\]](wrappers/nodejs.md)
* [Go Language Wrapper \[ALPHA\]](https://docs.seldon.io/projects/seldon-core/en/latest/go/go_wrapper_link.html)

### Ingress

* [Ambassador Ingress](routing/ambassador.md)
* [Istio Ingress](routing/istio.md)

### Production

* [Supported API Protocols](deployments/protocols.md)
* [CI/CD MLOps at Scale](integrations/cicd-mlops.md)
* [Metrics with Prometheus](integrations/analytics.md)
* [Payload Logging with ELK](integrations/logging.md)
* [Distributed Tracing with Jaeger](integrations/distributed-tracing.md)
* [Replica Scaling](deployments/scaling.md)
* [Budgeting Disruptions](deployments/disruption-budgets.md)
* [Custom Inference Servers](servers/custom.md)

### Advanced Inference

* [Model Explanations](integrations/explainers.md)
* [Outlier Detection](integrations/outlier_detection.md)
* [Routers (incl. Multi Armed Bandits)](routing/routers.md)

### Examples

* [Notebooks](notebooks/)

### Reference

* [Annotation-based Configuration](reference/annotations.md)
* [Benchmarking](reference/benchmarking.md)
* [General Availability](reference/ga.md)
* [Helm Charts](reference/helm_charts.md)
* [Images](https://docs.seldon.io/projects/seldon-core/en/latest/reference/images.html)
* [Logging & Log Level](reference/log_level.md)
* [Private Docker Registry](reference/private_registries.md)
* [Prediction APIs](https://docs.seldon.io/projects/seldon-core/en/latest/reference/apis/index.html)
* [Python API reference](https://docs.seldon.io/projects/seldon-core/en/latest/python/api/modules.html)
* [Release Highlights](https://docs.seldon.io/projects/seldon-core/en/latest/reference/release-highlights.html)
* [Seldon Deployment CRD ](reference/)(https://docs.seldon.io/projects/seldon-core/en/latest/reference/seldon-deployment.html)
* [Service Orchestrator](reference/svcorch.md)
* [Kubeflow](reference/kubeflow.md)

### Developer

* [Overview](developer/)
* [Contributing to Seldon Core](developer/contributing.md)
* [End to End Tests](developer/e2e.md)
* [Roadmap](developer/roadmap.md)
* [Build using private repo](developer/buid-using-private-repo.md)

## About the name "Seldon Core"

The name Seldon (ˈSɛldən) Core was inspired from [the Foundation Series (Sci-fi novels)](https://en.wikipedia.org/wiki/Foundation_series) where its premise consists of a mathematician called "Hari Seldon" who spends his life developing a theory of Psychohistory, a new and effective mathematical sociology which allows for the future to be predicted extremely accurate through long periods of time (across hundreds of thousands of years).

## Commercial Offerings

To learn more about our commercial offerings visit [https://www.seldon.io/](https://www.seldon.io/).

## License

[License](LICENSE/)
