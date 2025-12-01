# Alibi-Detect CloudEvents Server

Extends the [Seldon CloudEvents Server](https://github.com/SeldonIO/seldon-models/tree/master/servers/cloudevents) to allow Alibi Detect models to be loaded and events processed.

## Maintanence

In order to build and test integration with the Alibi Detect library, ensure to execute the following commands after pulling changes:

```bash

make dev_install
make test

```

This will install the Alibi Detect library in the development environment and run the tests.

There is also a possibility to quickly run sanity checks in this way:

```bash
make run-outlier-detector-tensorflow # to run the outlier detector server with tensorflow locally
make curl-detector-tensorflow # to send a test request to the server
make curl-tensorflow-outlier-detector-scores # to send a test request to get outlier scores

make run-outlier-detector-v2 # to run the outlier detector server with kfserving protocol locally
make curl-detector-v2 # to send a test request to the server
make curl-detector-v2-outlier # to send a test request to get outlier scores

make run-metrics-server # to run the metrics server locally
make curl-metrics-server-feedback # to submit a test feedback with truth
make curl-metrics-server-get # to get metrics from the server
```

In order to run the integration tests in the Kubernetes cluster, build the Docker image first: `make docker-build`.
Then, as part of the cluster and ecosystem setup in the `testing/scripts` folder, [this test](../../testing/scripts/test_alibi_detect_server.py) can be executed to validate the Alibi Detect server functionality.

## Deprecation Notice for GPU image (starting from Core 1.19.0 release)

GPU image for Alibi Detect server is deprecated and won't be maintained starting from Seldon Core 1.19.0 release.