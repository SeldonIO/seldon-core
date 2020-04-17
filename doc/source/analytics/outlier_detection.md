# Outlier Detection in Seldon Core

Machine learning models do not extrapolate well outside of the training data distribution. In order to trust and reliably act on model predictions, it is crucial to monitor the distribution of incoming requests via different types of detectors. Outlier detectors aim to flag individual instances which do not follow the original training distribution.

A [worked example with using the CIFAR10 task](../examples/outlier_cifar10.html) is available.

The general framework shown in this example is to use the Seldon Core payload logger to pass requests to components that process them asynchronously. The results can be passed onwards to alterting systems.

![Example architecture](analytics.png)