# Seldon Executor

This Go project replaces the Seldon Java Engine.  It is presently in development.

** Do not use in production **


## Functionality

The focus is to provide a smaller more efficient graph orchestror. It itends t provide:

 * http REST Seldon protocol server for Seldon graphs.
 * grpc Seldon protocol server for Seldon graphs.
 * Ability to handle other protocols in future, e.g. Tensorflow or NVIDIA TensorRT Server.

The REST and gRPC server are more restricted in that:

 * Only 1 is active and it assume your entire graph made up of REST or gRPC components. Mixing is not allowed.

The executor at present will not include functionality presently in the Java engine which will need to be provided elsewhere. Specifically:

 * No request logging
   * The roadmap will be to use the kfserving InferenceLogger we are proposing in that project and update the Seldon Deployment schema to allow that to be injected as needd.
 * No graph metrics
   * The roadmap will be to assume each graph component exposes their own Prometheus metrics
 * No meta data additions
   * It will be assumed this will be done by graph components.

To realise some of the above the Seldon Wrappers will need to be extended.

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
