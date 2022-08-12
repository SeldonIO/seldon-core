# SKLearn Server

If you have a trained SKLearn model saved as a pickle you can deploy it simply
using Seldon's prepackaged SKLearn server.

## Prerequisites

Seldon expects that your model has been saved using `joblib`, and it named as
`model.joblib`. 
Note that this is the [recommended approach to serialise
models](https://scikit-learn.org/stable/modules/model_persistence.html) by the
SKLearn project.

Note that, since we are using `joblib`, it's important that your trained model
matches the framework version expected in the inference server.

The expected versions in the latest SKLearn pre-packaged server are as follows:

| Package | Version |
| ------ | ----- |
| `scikit-learn` | `0.24.2` |

To check compatibility requirements for older versions of Seldon Core you can
see the [compatibility table below](#version-compatibility).

## Usage

To use the pre-packaged SKLearn server, it's enough to declare `SKLEARN_SERVER`
as the `implementation` for your model.
For example, for a saved Iris prediction model, you could do:

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.15.0-dev/sklearn/iris
      name: classifier
    name: default
    replicas: 1

```

You can try a similar example in [this worked
notebook](../examples/server_examples.html).

### Sklearn inference method

By default the server will call `predict_proba` on your loaded model/pipeline.
If you wish for it to call `predict` instead you can pass a parameter `method`
and set it to `predict`.
For example:

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris-predict
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.15.0-dev/sklearn/iris
      name: classifier
      parameters:
        - name: method
          type: STRING
          value: predict
    name: default
    replicas: 1
```

Acceptable values for the `method` parameter are `predict`, `predict_proba`,
`decision_function`.


## V2 protocol

The SKLearn server can also be used to expose an API compatible with the [V2
V2 Protocol](../graph/protocols.md#v2-protocol).
Note that, under the hood, it will use the [Seldon
MLServer](https://github.com/SeldonIO/MLServer) runtime.

In order to enable support for the V2 protocol, it's enough to
specify the `protocol` of the `SeldonDeployment` to use `v2`.
For example,

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris-predict
  protocol: v2 # Activate the V2 protocol
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.15.0-dev/sklearn/iris
      name: classifier
      parameters:
        - name: method
          type: STRING
          value: predict
    name: default
```

You can try a similar example in [this worked
notebook](../examples/server_examples.html).

## Version compatibility

The version of SKLearn used by the pre-packaged inference server will depend on
the installed version of Seldon Core.
In particular, 

| Seldon Version | SKLearn Version |
| -------------- | --------------- |
| `>=1.3`          | `0.23.2`          |
| `<1.3` (latest `1.2.3`)          | `0.20.3`          |

Note that using a different version of SKLearn at training and inference time
can cause unexpected issues when it comes to serving.

### Using an older version

If you wish to use an older image of the SKLearn inference server, you can
override the used image in the `componentSpecs`.
For example, to use version `1.2.3` of the SKLearn server you could do:

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - componentSpecs:
    - spec:
       containers:
       - name: classifier
         image: seldonio/sklearnserver_rest:1.2.3
    graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/v1.15.0-dev/sklearn/iris
      name: classifier
    name: default
    replicas: 1
    svcOrchSpec: 
      env: 
      - name: SELDON_LOG_LEVEL
        value: DEBUG
```

### Using a different version of SKLearn

If you wish to use an unsupported version of SKLearn, you can extend the
existing SKLearn server to build your own. 
In particular, you could extend the code in the
[`servers/sklearnserver`](https://github.com/SeldonIO/seldon-core/tree/master/servers/sklearnserver)
folder to build a custom image.
This image used for the `SKLEARN_SERVER` implementation can then be overridden
in the `componentSpecs`.

Note that you can also change the image used globally for the SKLearn server by
editing the [`seldon-config` configmap](custom.md).
This change would apply to all `SeldonDeployments` in your cluster leveraging
the `SKLEARN_SERVER` implementation.
For example, you could add the following to the configmap:

```yaml
  SKLEARN_SERVER:
    grpc:
      defaultImageVersion: "1.2.3"
      image: seldonio/sklearnserver_grpc
    rest:
      defaultImageVersion: "1.2.3"
      image: seldonio/sklearnserver_rest
```
