# Seldon Core Python Package

Seldon Core has a python package `seldon_core` available on PyPI. The package makes it easier to work with Seldon Core if you are using python and is the basis of the Python S2I wrapper. The module provides:

 * `seldon-core-microservice` executable to serve microservice components in Seldon Core. This is used by the Python Wrapper for Seldon Core.
 * `seldon-core-microservice-tester` executable to test running Seldon Core microservices over REST or gRPC.
 * `seldon-core-api-tester` executable to test the external API for running Seldon Deployment inference graphs over REST or gRPC.
 * `seldon_core.seldon_client` library. Core reference API module to call Seldon Core services (internal microservices or the external API). This is used by the testing executable and can be used by users to build their own clients to Seldon Core in Python.

## Install

Install from PyPI with:

```bash
pip install seldon-core
```

### Tensorflow support

Seldon Core adds optional support to send a `TFTensor` as your prediction
input.
However, most users will prefer to send a `numpy` array, string, binary or JSON
input instead.
Therefore, in order to avoid including the `tensorflow` dependency on
installations where the `TFTensor` support won't be necessary, it isn't
installed it by default.

To include the optional Tensorflow support, you can install `seldon_core` as:

```bash
pip install seldon_core[tensorflow]
```

### Google Cloud Storage support

As part of the options to store your trained model, Seldon Core adds optional
support to fetch them from GCS (Google Cloud Storage).
We are aware that users will usually only require one of the storage backends.
Therefore, to avoid bloating the `seldon_core` package, we don't install the
GCS dependencies by default.

To include the optional GCS support, you can install `seldon_core` as:

```bash
pip install seldon_core[gcs]
```

### Install all extra dependencies

If you want to install `seldon_core` with all its extra dependencies, you can
do so as:

```bash
pip install seldon_core[all]
```

Keep in mind that this will include some dependencies which may not be used.
Therefore, unless necessary, we recommend most users to install the default
distribution of `seldon_core` as [documented above](#install).

## Seldon Core Microservices

Seldon allows you to easily take your runtime inference code and create a Docker container that can be managed by Seldon Core. Follow the [S2I instructions](../wrappers/python.md) to wrap your code.

You can also create your own image and utilise the `seldon-core-microservice` executable to run your model code.

## Testing Seldon Core Microservices

To test your microservice standalone or your running Seldon Deployment inside Kubernetes you can follow the [API testing docs](../workflow/api-testing.md).


## Seldon Core Python API Client

The python package contains a module that provides a reference python client for the internal Seldon Core microservice API and the external APIs. More specifically it provides:

 * Internal microservice API
    * Make REST or gRPC calls
    * Test all methods: `predict`, `transform-input`, `transform-output`, `route`, `aggregate`
    * Provide a numpy array, binary data or string data as payload or get random data generated as payload for given shape
    * Send data as tensor, TFTensor or ndarray
 * External API
    * Make REST or gRPC calls
    * Call the API via Ambassador, Istio or Seldon's OAUTH API gateway.
    * Test `predict` or `feedback` endpoints
    * Provide a numpy array, binary data or string data as payload or get random data generated as payload for given shape
    * Send data as tensor, TFTensor or ndarray

Basic usage of the client is to create a `SeldonClient` object first. For example for a Seldon Deployment called "mymodel" running in the namespace `seldon` with Ambassador endpoint at "localhost:8003" (i.e., via port-forwarding):

```python
from seldon_core.seldon_client import SeldonClient
sc = SeldonClient(deployment_name="mymodel",namespace="seldon", gateway_endpoint="localhost:8003")
```

Then make calls of various types. For example, to make a random prediction  via the Ambassador gateway using REST:

```python
r = sc.predict(gateway="ambassador",transport="rest")
print(r)
```

Examples of using the `seldon_client` module can be found in the [example notebook](../examples/helm_examples.html).

The API docs can be found [here](api/seldon_core.html#module-seldon_core.seldon_client).

