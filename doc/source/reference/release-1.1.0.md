# Seldon Core Release 1.1.0

A summary of the main contributions to the [Seldon Core release 1.1.0](https://github.com/SeldonIO/seldon-core/releases/tag/v1.1.0).

## Tensorflow Protocol Supported

We now support Tensorflow Protocol natively via REST and gRPC. As Seldon provides a full inference graph there are some [restrictions and extensions to how the protocol can be used](../graph/protocols.html). 

## New Service Orchestrator

The Seldon Service Orchestrator manages the request and response flow through the defined inference graph. A new simplified GoLang executor has been released which provides core management of the request/response flow as well as tracing, metrics and payload logging. It has been designed to allow easy addition of new dataplanes over REST or gRPC. Further details can be found [here](../graph/svcorch.md).

## Outlier and Drift Detection Examples

Once you have deployed a model using Seldon Core its key to ensure its running as expected. Two key areas for this are outlier detection and drift detection. Outlier detection identified individuals requests that seem to fall out of the expected distribution of features the model was trained on. These need to be identified as they will likely cause unreliable predictions from the deployed model. Drift detecton identified when the observed distribution of features over some time period has changed to what the model was trained on. This is important so that data scientists can be informed to updates and train a new model on more recent data.

We provide two examples in out docs:

 * [Outlier detection on a CIFAR10 image model](../analytics/outlier_detection.html).
 * [Drift detection on a CIFAR10 image model](../analytics/drift_detection.html).

## Scale CRDs

We now provide the ability to scale SeldonDeployments via the `kubectl scale` command. Read further on  [scaling and setting the correct replicas for each part of your inference graph](../graph/scaling.html). 

## RedHat Community Operator

We have released onto the [OperatorHub](https://operatorhub.io/operator/seldon-operator) and RedHat Community operators so you can now easily install Seldon Core via these distribution channels.


