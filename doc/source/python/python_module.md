# Seldon Core Python Package

Seldon Core has a python package `seldon-core` available on PyPI. The package makes it easier to work with Seldon Core if you are using python and is the basis of the Python S2I wrapper. The module provides:

 * `seldon-core-microservice` executable to serve microservice components in Seldon Core. This is used by the Python Wrapper for Seldon Core.
 * `seldon_core.seldon_client` library. Core reference API module to call Seldon Core services (internal microservices or the external API). This is used by the testing executable and can be used by users to build their own clients to Seldon Core in Python.

## Install

Install from PyPI with:

```bash
$ pip install seldon-core
```

### Tensorflow support

Seldon Core adds optional support to send a `TFTensor` as your prediction input.
However, most users will prefer to send a `numpy` array, string, binary or JSON input instead.
Therefore, in order to avoid including the `tensorflow` dependency on installations where the `TFTensor` support won't be necessary, it isn't installed it by default.

To include the optional `TFTensor` support, you can install `seldon-core` as:

```bash
$ pip install seldon-core[tensorflow]
```

### Google Cloud Storage support

As part of the options to store your trained model, Seldon Core adds optional
support to fetch them from GCS (Google Cloud Storage).
We are aware that users will usually only require one of the storage backends.
Therefore, to avoid bloating the `seldon-core` package, we don't install the
GCS dependencies by default.

To include the optional GCS support, you can install `seldon-core` as:

```bash
$ pip install seldon-core[gcs]
```

We are currently looking into options to replace the multiple cloud storage
libraries that `seldon-core` requires for a single multi-cloud one.
This discussion is currently open on [issue #1028](https://github.com/SeldonIO/seldon-core/issues/1028).
Feedback and suggestions are welcome!

### Azure Blob Storage support

As part of the options to store your trained model, Seldon Core adds optional
support to fetch them from Azure Blob Storage.
We are aware that users will usually only require one of the storage backends.
Therefore, to avoid bloating the `seldon-core` package, the Azure Blob Storage
dependencies are not installed by default.

To include the optional Azure support, you can install `seldon-core` as:

```bash
$ pip install seldon-core[azure]
```

### Install all extra dependencies

If you want to install `seldon-core` with all its extra dependencies, you can
do so as:

```bash
$ pip install seldon-core[all]
```

Keep in mind that this will include some dependencies which may not be used.
Therefore, unless necessary, we recommend most users to install the default
distribution of `seldon-core` as [documented above](#install).

## Seldon Core Microservices

Seldon allows you to easily take your runtime inference code and create a Docker container that can be managed by Seldon Core. Follow the [S2I instructions](../wrappers/python.md) to wrap your code.

You can also create your own image and utilise the `seldon-core-microservice` executable to run your model code.


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

The API docs can be found [here](./api/seldon_core.html#module-seldon_core.seldon_client).

## Troubleshooting

If you experience problems after installing `seldon-core`, here are some tips
to diagnose the issue.

### ImportError: cannot import name 'BlockBlobService'

The library we use to support Azure Blob Storage [released an
update](https://github.com/Azure/azure-storage-python/issues/640) which
contains breaking changes with previous versions.
This update breaks versions of `seldon-core` below or equal to `0.5.0` but it
shouldn't affect users on version `0.5.0.2` and above.
If you are facing this issue, you should see a stacktrace similar to the one
below:

```pytb
.../seldon_core/storage.py in <module>
     23 import re
     24 from urllib.parse import urlparse
---> 25 from azure.storage.blob import BlockBlobService
     26 from minio import Minio
     27 from seldon_core.imports_helper import _GCS_PRESENT

ImportError: cannot import name 'BlockBlobService'
```

The recommended workaround is to update `seldon-core` to version `0.5.0.2` or
above.
Alternatively, if you can't upgrade to a more recent version, the following
also works:

```bash
$ pip install azure-storage-blob==2.1.0 seldon-core
```
