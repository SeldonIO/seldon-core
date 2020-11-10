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

## Next Steps

[Create your python inference class](python_component.md)


