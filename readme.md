# Seldon Core: Blazing Fast, Industry-Ready ML
An open source platform to deploy your machine learning models on Kubernetes at massive scale.

![API](./doc/source/images/core-logo-small.png)


# Overview

Seldon core converts your ML models (Tensorflow, Pytorch, H2o, etc.) or language wrappers (Python, Java, etc.) into production REST/GRPC microservices.

Seldon handles scaling to thousands of production machine learning models and provides advanced machine learning capabilities out of the box including Advanced Metrics, Request Logging, Explainers, Outlier Detectors, A/B Tests, Canaries and more.

* Read the [Seldon Core Documentation](https://docs.seldon.io/projects/seldon-core/en/latest/)
* Get started with [Seldon Core Notebook Examples](https://docs.seldon.io/projects/seldon-core/en/latest/examples/notebooks.html)
* Join our [community Slack](https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLTJlNjQ1NTE5Y2MzMWIwMGUzYjNmZGFjZjUxODU5Y2EyMDY0M2U3ZmRiYTBkOTRjMzZhZjA4NjJkNDkxZTA2YmU)
* Learn how you can [start contributing](https://docs.seldon.io/projects/seldon-core/en/latest/developer/contributing.html)
 
![API](./doc/source/images/seldon-core-high-level.jpg)

## Details

With over 2M installs, Seldon Core is used across organisations to manage large scale deployment of machine learning models, and key benefits include:

 * Cloud agnostic and runs production workflows on AWS EKS, Azure AKS, Google GKE, Alicloud, Digital Ocean and Openshift.
 * Convert your machine learning models into full fledged microservices using our language wrappers or pre-packaged inference servers
 * Create powerful and rich inference graphs made up of predictors, transformers, routers, combiners, and more.
 * Provide a standardised serving layer across models from heterogeneous toolkits and languages

# Getting Started



## So Why Seldon?

 * Lets you focus on your model by [making it easy to serve on kubernetes](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/README.html)
 * The same workflow and base API for a range of toolkits such as Sklearn, tensorflow and R
 * Out of the box best-practices for logging, tracing and base metrics, applicable to all models across toolkits
 * Support for deployment strategies such as running A/B test and canaries
 * Inferences graphs for microservice-based serving strategies such as multi-armed bandits or pre-processing

## Community

 * [Join our Slack Channel](https://join.slack.com/t/seldondev/shared_invite/enQtMzA2Mzk1Mzg0NjczLTJlNjQ1NTE5Y2MzMWIwMGUzYjNmZGFjZjUxODU5Y2EyMDY0M2U3ZmRiYTBkOTRjMzZhZjA4NjJkNDkxZTA2YmU)

## Details

With over 2M installs, Seldon Core is used across organisations to manage large scale deployment of machine learning models, and key benefits include:

 * Cloud agnostic and runs production workflows on AWS EKS, Azure AKS, Google GKE, Alicloud, Digital Ocean and Openshift
 * Get metrics and ensure proper governance and compliance for your running machine learning models.
 * Create powerful inference graphs made up of multiple components.
 * Provide a consistent serving layer for models built using heterogeneous ML toolkits.

# Getting Started


## Seldon Core in Action (High Level Examples)

### Using Prepackaged Inference Servers

You can leverage our optimised pre-packaged inference servers which are set up to load binaries from specific frameworks. We currently have pre-packaged inference servers for Tensorflow, XGBoost, SKlearn and MLFlow models. Learn everything about pre-packaged inference servers [in its documentation section]().

The high level steps involved in using a prepackaged model server is as follows:

1. Export your model binaries and/or artifacts:

```python
>>my_sklearn_model.train(...)
>>pickle.dumps(my_sklearn_model, "model.pickle")

[Created file at /mypath/model.pickle]
```

2. Test locally with our prepackaged model servers

```console
$ sc --model-server SKLEARN --file /mypath/model.pickle

Listening on port 8080...

$ curl -X POST localhost:8080/api/v1.0/predictions \
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

3. Upload your model to an object store

```
gs://seldon-models/sklearn/iris/model.pickle
```

4. Deploy to Kubernetes

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

5. Send a request in Kubernetes cluster

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

#### Using our language wrappers

1. Export your model binaries and/or artifacts:

```python
>> my_sklearn_model.train(...)
>> pickle.dumps(my_sklearn_model, "model.pickle")

[Created file at /mypath/model.pickle]
```

2. Create a wrapper class `Model.py`
```python
class Model:
    def __init__(self):
        self._model = pickle.loads("model.pickle")

    def predict(self, X):
        output = self._model(X)
        return output
```

3. Use the Seldon tools to containerise your model

```console
s2i build . seldonio/seldon-core-s2i-python3:0.18 sklearn_iris:0.1 
```

4. Test model locally

```console
$ docker run -p 8000:8000 --rm step_one:0.1 

Listening on port 8080...

$ curl -X POST localhost:8080/api/v1.0/predictions \
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

4. Deploy to Kubernetes

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

5. Send a request in Kubernetes cluster

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

