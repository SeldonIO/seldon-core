# Seldon Core: Blazing Fast, Industry-Ready ML
An open source platform to deploy your machine learning models on Kubernetes at massive scale.

![](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/core-logo-small.png)


## Overview

Seldon core converts your ML models (Tensorflow, Pytorch, H2o, etc.) or language wrappers (Python, Java, etc.) into production REST/GRPC microservices.

Seldon handles scaling to thousands of production machine learning models and provides advanced machine learning capabilities out of the box including Advanced Metrics, Request Logging, Explainers, Outlier Detectors, A/B Tests, Canaries and more.

* Read the [Seldon Core Documentation](https://docs.seldon.io/projects/seldon-core/en/latest/)
* Join our [community Slack](https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLTJlNjQ1NTE5Y2MzMWIwMGUzYjNmZGFjZjUxODU5Y2EyMDY0M2U3ZmRiYTBkOTRjMzZhZjA4NjJkNDkxZTA2YmU) to ask any questions
* Get started with [Seldon Core Notebook Examples](https://docs.seldon.io/projects/seldon-core/en/latest/examples/notebooks.html)
* Join our fortnightly [online community calls](https://docs.seldon.io/projects/seldon-core/en/latest/developer/community.html)
* Learn how you can [start contributing](https://docs.seldon.io/projects/seldon-core/en/latest/developer/contributing.html)
* Check out [Blogs](https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/blogs.html) that dive into Seldon Core components
* Watch some of the [Videos and Talks](https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/videos.html) using Seldon Core

![](https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/seldon-core-high-level.jpg)

### High Level Features

With over 2M installs, Seldon Core is used across organisations to manage large scale deployment of machine learning models, and key benefits include:

 * Cloud agnostic and tested on AWS EKS, Azure AKS, Google GKE, Alicloud, Digital Ocean and Openshift.
 * Easy way to containerise ML models using our language wrappers or pre-packaged inference servers.
 * Powerful and rich inference graphs made out of predictors, transformers, routers, combiners, and more.
 * A standardised serving layer across models from heterogeneous toolkits and languages.
 * Advanced and customisable metrics with integration to Prometheus and Grafana.
 * Full auditability through model input-output request logging integration with Elasticsearch.
 * Microservice tracing through integration to Jaeger for insights on latency across microservice hops.


## Getting Started

Deploying your models using Seldon Core is simplified through our pre-packaged inference servers and language wrappers. Below you can see how you can deploy our "hello world Iris" example. You can see more details on these workflows in our [Documentation Quickstart](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/quickstart.html).

### Install Seldon Core

Quick install using Helm 3 (you can also use Kustomize):

```
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
gs://seldon-models/sklearn/iris/model.pickle
```

We then can deploy this model with Seldon Core to our Kubernetes cluster using the pre-packaged model server for scikit-learn (SKLEARN_SERVER) by running the `kubectl apply` command below:

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
  - graph:
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1
END
```

#### Send API requests to your deployed model

Once our model is deployed we can send requests using either the REST or GRPC endpoint:

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

### Deploy your custom model using language wrappers

For more custom deep learning and machine learning use-cases which have custom dependencies (such as 3rd party libraries, operating system binaries or even external systems), we can use any of the Seldon Core language wrappers.

You only have to write a class wrapper that exposes the logic of your model; for example in Python we have create a file `Model.py`:

```python
import pickle
class Model:
    def __init__(self):
        self._model = pickle.loads( open("model.pickle", "rb") )

    def predict(self, X):
        output = self._model(X)
        return output
```

We can now containerize our class file using the [Seldon Core s2i utils](https://docs.seldon.io/projects/seldon-core/en/latest/wrappers/s2i.html) to produce the `sklearn_iris` image:

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
  - graph:
      name: classifier
    name: default
    replicas: 1
END
```


#### Send API requests to your deployed model

Once our model is deployed we can send requests using either the REST or GRPC endpoint:

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

<table>
  <tr valign="top">
    <td width="50%" >
        <a href="https://docs.seldon.io/projects/seldon-core/en/latest/analytics/analytics.html">
            <br>
            <b>Standard and custom metrics with prometheus</b>
            <br>
            <br>
            <img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/analytics/dashboard.png">
        </a>
    </td>
    <td width="50%">
        <a href="https://docs.seldon.io/projects/seldon-core/en/latest/analytics/logging.html">
            <br>
            <b>Full audit trails with ELK request logging</b>
            <br>
            <br>
            <img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/kibana-custom-search.png">
        </a>
    </td>
  </tr>
  <tr valign="top">
    <td width="50%">
        <a href="https://docs.seldon.io/projects/seldon-core/en/latest/analytics/explainers.html">
            <br>
            <b>Explainers for Machine Learning Interpretability</b>
            <br>
            <br>
            <img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/anchors.jpg">
        </a>
    </td>
    <td width="50%">
        <a href="https://docs.seldon.io/projects/seldon-core/en/latest/analytics/outlier_detection.html">
            <br>
            <b>Outlier and Adversarial Detectors for Monitoring</b>
            <br>
            <br>
            <img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/adversarial-attack.png">
        </a>
    </td>
  </tr>
  <tr valign="top">
    <td width="50%">
        <a href="https://docs.seldon.io/projects/seldon-core/en/latest/analytics/cicd-mlops.html">
            <br>
            <b>CI/CD for MLOps at Massive Scale</b>
            <br>
            <br>
            <img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/images/cicd-seldon.jpg">
        </a>
    </td>
    <td width="50%">
        <a href="https://docs.seldon.io/projects/seldon-core/en/latest/graph/distributed-tracing.html">
            <br>
            <b>Distributed tracing for performance monitoring</b>
            <br>
            <br>
            <img src="https://raw.githubusercontent.com/SeldonIO/seldon-core/master/doc/source/graph/jaeger-ui-rest-example.png">
        </a>
    </td>
  </tr>
</table>


## Where to go from here

### Getting Started

* [Overview ](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/github-readme.html)
* [Quickstart Guide ](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/quickstart.html)
* [Install Seldon Core on Kubernetes ](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html)
* [Join the Community ](https://docs.seldon.io/projects/seldon-core/en/latest/developer/community.html)

### Seldon Core Deep Dive
  
* [Detailed Installation Parameters ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/helm.html)
* [Pre-packaged Inreference Servers ](https://docs.seldon.io/projects/seldon-core/en/latest/servers/overview.html)
* [Language Wrappers for Custom Models ](https://docs.seldon.io/projects/seldon-core/en/latest/wrappers/language_wrappers.html)
* [Create your Inference Graph ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/inference-graph.html)
* [Deploy your Model  ](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/deploying.html)
* [Testing your Model Endpoints  ](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/serving.html)
* [Troubleshooting guide ](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/troubleshooting.html)
* [Usage reporting ](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/usage-reporting.html)
* [Upgrading ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/upgrading.html)
* [Changelog ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/changelog.html)

### Pre-Packaged Inference Servers
	     
* [MLflow Server ](https://docs.seldon.io/projects/seldon-core/en/latest/servers/mlflow.html)
* [SKLearn server ](https://docs.seldon.io/projects/seldon-core/en/latest/servers/sklearn.html)
* [Tensorflow Serving ](https://docs.seldon.io/projects/seldon-core/en/latest/servers/tensorflow.html)
* [XGBoost server ](https://docs.seldon.io/projects/seldon-core/en/latest/servers/xgboost.html)
   
### Language Wrappers (Production)

* [Python Language Wrapper [Production] ](https://docs.seldon.io/projects/seldon-core/en/latest/python/index.html)

### Language Wrappers (Incubating)

* [Java Language Wrapper [Incubating] ](https://docs.seldon.io/projects/seldon-core/en/latest/java/README.html)
* [R Language Wrapper [ALPHA] ](https://docs.seldon.io/projects/seldon-core/en/latest/R/README.html)
* [NodeJS Language Wrapper [ALPHA] ](https://docs.seldon.io/projects/seldon-core/en/latest/nodejs/README.html)
* [Go Language Wrapper [ALPHA] ](https://docs.seldon.io/projects/seldon-core/en/latest/go/go_wrapper_link.html)

### Ingress

* [Ambassador Ingress ](https://docs.seldon.io/projects/seldon-core/en/latest/ingress/ambassador.html)
* [Istio Ingress ](https://docs.seldon.io/projects/seldon-core/en/latest/ingress/istio.html)
   
### Production

* [Supported API Protocols ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/protocols.html)
* [CI/CD MLOps at Scale ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/cicd-mlops.html)
* [Metrics with Prometheus ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/analytics.html)
* [Payload Logging with ELK ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/logging.html)
* [Distributed Tracing with Jaeger ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/distributed-tracing.html)
* [Autoscaling in Kubernetes ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/autoscaling.html)
      
### Advanced Inference

* [Model Explanations ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/explainers.html)
* [Outlier Detection ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/outlier_detection.html)
* [Routers (incl. Multi Armed Bandits)  ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/routers.html)   
   
### Examples

* [Notebooks ](https://docs.seldon.io/projects/seldon-core/en/latest/examples/notebooks.html)
* [Articles/Blogs ](https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/blogs.html)
* [Videos ](https://docs.seldon.io/projects/seldon-core/en/latest/tutorials/videos.html)

### Reference

* [Annotation-based Configuration ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/annotations.html)   	     
* [AWS Marketplace Install ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/aws-mp-install.html)
* [Benchmarking ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/benchmarking.html)   
* [General Availability ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/ga.html)
* [Helm Charts ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/helm_charts.html)
* [Images ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/images.html)
* [Logging & Log Level ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/log_level.html)   
* [Private Docker Registry ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/private_registries.html)   
* [Prediction APIs ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/apis/index.html)   
* [Python API reference ](https://docs.seldon.io/projects/seldon-core/en/latest/python/api/modules.html)
* [Release Highlights ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/release-highlights.html)   
* [Seldon Deployment CRD ](https://docs.seldon.io/projects/seldon-core/en/latest/reference/seldon-deployment.html)   
* [Service Orchestrator ](https://docs.seldon.io/projects/seldon-core/en/latest/graph/svcorch.html)
* [Kubeflow ](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/kubeflow.html)

### Developer

* [Overview ](https://docs.seldon.io/projects/seldon-core/en/latest/developer/readme.html)
* [Contributing to Seldon Core ](https://docs.seldon.io/projects/seldon-core/en/latest/developer/contributing.html)
* [End to End Tests ](https://docs.seldon.io/projects/seldon-core/en/latest/developer/e2e.html)
* [Build using private repo ](https://docs.seldon.io/projects/seldon-core/en/latest/developer/build-using-private-repo.html)



## About the name "Seldon Core"

The name Seldon (ˈSɛldən) Core was inspired from [the Foundation Series (Scifi Novel)](https://en.wikipedia.org/wiki/Foundation_series) where it's premise consists of a mathematician called "Hari Seldon" who spends his life developing a theory of Psychohistory, a new and effective mathematical sociology which allows for the future to be predicted extremely accurate through long periods of time (across hundreds of thousands of years).

