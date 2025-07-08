# Scikit-Learn Iris Model using customData

* Wrap a scikit-learn python model for use as a prediction microservice in seldon-core
    * Run locally on Docker to test
    * Deploy on seldon-core running on a Kubernetes cluster

## Dependencies

* [s2i](https://github.com/openshift/source-to-image)
* Seldon Core v1.0.3+ installed
* `pip install sklearn seldon-core protobuf grpcio`

## Train locally


```python
import os

import numpy as np
from sklearn import datasets
from sklearn.externals import joblib
from sklearn.linear_model import LogisticRegression
from sklearn.pipeline import Pipeline


def main():
    clf = LogisticRegression()
    p = Pipeline([("clf", clf)])
    print("Training model...")
    p.fit(X, y)
    print("Model trained!")

    filename_p = "IrisClassifier.sav"
    print("Saving model in %s" % filename_p)
    joblib.dump(p, filename_p)
    print("Model saved!")


if __name__ == "__main__":
    print("Loading iris data set...")
    iris = datasets.load_iris()
    X, y = iris.data, iris.target
    print("Dataset loaded!")
    main()
```

## Custom Protobuf Specification

First, we'll need to define our custom protobuf specification so that it can be leveraged.


```python
%%writefile iris.proto

syntax = "proto3";

package iris;

message IrisPredictRequest {
    float sepal_length = 1;
    float sepal_width = 2;
    float petal_length = 3;
    float petal_width = 4;
}

message IrisPredictResponse {
    float setosa = 1;
    float versicolor = 2;
    float virginica = 3;
}
```

## Custom Protobuf Compilation

We will need to compile our custom protobuf for python so that we can unpack the `customData` field passed to our `predict` method later on.


```python
!python -m grpc.tools.protoc --python_out=./ --proto_path=. iris.proto
```

## gRPC test

Wrap model using s2i


```python
!s2i build . seldonio/seldon-core-s2i-python37-ubi8:1.7.0-dev seldonio/sklearn-iris-customdata:0.1
```

Serve the model locally


```python
!docker run --name "iris_predictor" -d --rm -p 5000:5000 seldonio/sklearn-iris-customdata:0.1
```

Test using custom protobuf payload


```python
import grpc
from iris_pb2 import IrisPredictRequest, IrisPredictResponse

from seldon_core.proto import prediction_pb2, prediction_pb2_grpc

channel = grpc.insecure_channel("localhost:5000")
stub = prediction_pb2_grpc.ModelStub(channel)

iris_request = IrisPredictRequest(
    sepal_length=7.233, sepal_width=4.652, petal_length=7.39, petal_width=0.324
)

seldon_request = prediction_pb2.SeldonMessage()
seldon_request.customData.Pack(iris_request)

response = stub.Predict(seldon_request)

iris_response = IrisPredictResponse()
response.customData.Unpack(iris_response)

print(iris_response)
```

Stop serving model


```python
!docker rm iris_predictor --force
```

## Setup Seldon Core

Use the [setup notebook](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/seldon_core_setup.ipynb) to setup Seldon Core with an ingress - either Ambassador or Istio

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

* Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
* Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:80`


```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

## Deploy your Seldon Model

We first create a configuration file:


```python
%%writefile sklearn_iris_customdata_deployment.yaml

apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-deployment-example
spec:
  name: sklearn-iris-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: groszewn/sklearn-iris-customdata:0.1
          imagePullPolicy: IfNotPresent
          name: sklearn-iris-classifier
    graph:
      children: []
      endpoint:
        type: GRPC
      name: sklearn-iris-classifier
      type: MODEL
    name: sklearn-iris-predictor
    replicas: 1
```

### Run the model in our cluster

Apply the Seldon Deployment configuration file we just created


```python
!kubectl create -f sklearn_iris_customdata_deployment.yaml
```

### Check that the model has been deployed


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-deployment-example -o jsonpath='{.items[0].metadata.name}')
```

## Test by sending prediction calls

`IrisPredictRequest` sent via the `customData` field.


```python
iris_request = IrisPredictRequest(
    sepal_length=7.233, sepal_width=4.652, petal_length=7.39, petal_width=0.324
)

seldon_request = prediction_pb2.SeldonMessage()
seldon_request.customData.Pack(iris_request)

channel = grpc.insecure_channel("localhost:8003")
stub = prediction_pb2_grpc.SeldonStub(channel)

metadata = [("seldon", "seldon-deployment-example"), ("namespace", "seldon")]

response = stub.Predict(request=seldon_request, metadata=metadata)

iris_response = IrisPredictResponse()
response.customData.Unpack(iris_response)

print(iris_response)
```

### Cleanup our deployment


```python
!kubectl delete -f sklearn_iris_customdata_deployment.yaml
```
