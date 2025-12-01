# Alibi-Detect CloudEvents Server

Extends the [Seldon CloudEvents Server](https://github.com/SeldonIO/seldon-models/tree/master/servers/cloudevents) to allow Alibi Detect models to be loaded and events processed.

## Maintanence

In order to build and test integration with the Alibi Detect library, ensure to execute the following commands after pulling changes:

```bash

make dev_install
make test


```

This will install the Alibi Detect library in the development environment and run the tests.

In order to run the integration tests in the Kubernetes cluster, build the Docker image first: `make docker-build`.
Then, as part of the cluster and ecosystem setup in the `testing/scripts` folder, [this test](../../testing/scripts/test_alibi_detect_server.py) can be executed to validate the Alibi Detect server functionality.

## Deprecation Notice for GPU image (starting from Core 1.19.0 release)

GPU image for Alibi Detect server is deprecated and won't be maintained starting from Seldon Core 1.19.0 release.