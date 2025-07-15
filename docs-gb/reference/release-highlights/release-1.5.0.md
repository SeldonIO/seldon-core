# Seldon Core Release 1.5.0

A summary of the main contributions to the [Seldon Core release 1.5.0](https://github.com/SeldonIO/seldon-core/releases/tag/v1.5.0).

## Inference endpoints for REST and gRPC by default

We now expose both REST and gRPC endpoints for all inference graphs by default. This means the graph definitions no longer need to contain `endpoint.type`. This has been made possible by updating our python wrapper to allow both REST and gRPC endpoints to be exposed along with all the prepackaged servers. So in the past you would have a resource like below for a gRPC model:

```
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: grpc-seldon
spec:
  name: grpcseldon
  protocol: seldon
  transport: grpc
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier_grpc:1.3
          name: classifier
    graph:
      name: classifier
      type: MODEL
      endpoint:
        type: GRPC
    name: model
    replicas: 1
```

From 1.5 the same resource would look like: 

```
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: example-seldon
spec:
  protocol: seldon
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.5.0
          name: classifier
    graph:
      name: classifier
      type: MODEL
    name: model
    replicas: 1
```

This assumes `mock_classifier:1.5.0` has been wrapped with the 1.5.0 python wrapper. For models wrapped with older versions of Seldon Core the REST or gRPC endpoint will still continue to function. See [upgrading](upgrading.md) for details.

Istio and Amabssador configurations have been updated to allow both REST and gRPC configurations.

## Updated Drift, Outlier and Explanations with Alibi

Our [CIFAR10 outlier detection example](../examples/drift_cifar10.html) and [CIFAR10 drift detection example](../examples/drift_cifar10.html) have been updated and now requires KNative 0.18 with a compatible istio (tested on 1.7.3). This utilizes the v1 resources of KNative Eventing to show off asnychronous outlier and drift detection on images.

Our explanations examples using [Alibi:Explain](https://github.com/SeldonIO/alibi) have been updated to the latest 0.4.3 release.

## Other highlights

 * Our batch processor has been updated to allow feedback requests to be sent to allow accuracy and other metrics to be tracked for deployed models.
 * Our [KEDA autoscaling example](../examples/keda.html) has been updated to use the stable v1 release of KEDA. 



