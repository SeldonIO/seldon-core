# Seldon Executor

This Go project replaces the Seldon Java Engine.  It is presently in development.

## Functionality

The focus is to provide a smaller more efficient graph orchestror.

 * REST and gRPC for Seldon and Tensorflow protocols. Easily extendable to other protocols.
 * Logging of request and or response payloads to arbitrary URLs with CloudEvents
 * Tracing for REST and gRPC
 * Prometheus metrics for REST and gRPC

Changes to existing service orchestrator

 * All components must be REST or gRPC in agraph. No mixing.
 * Not meta data additions to payloads are carried out by the executor.


## Testing

You can choose to use this executor by adding the annotation `seldon.io/executor: "true"`. This annotation will be active until this project progresses from incubating status.

An example is shown below:

```JSON
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: seldon-model
spec:
  annotations:
    seldon.io/executor: "true"
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_rest:1.3
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```

## Development

We assume:

 * Go 1.13
 * golangci-lint v1.35.2

For linting the `golangci-lint` binary is required. To install, follow the [official install instructions](https://golangci-lint.run/usage/install/) by running:
```shell
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.35.2
```
