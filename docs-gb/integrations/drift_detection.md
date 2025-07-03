# Drift Detection in Seldon Core

Machine learning models do not extrapolate well outside of the training data distribution. In order to trust and reliably act on model predictions, it is crucial to monitor the distribution of incoming requests via different types of detectors.  Drift detectors check when the distribution of the incoming requests is diverging from a reference distribution such as that of the training data. If data drift occurs, the model performance can deteriorate and it should be retrained.


[Drift Detection example using CIFAR10](../examples/drift_cifar10.html).

The general framework shown in this example is to use the Seldon Core payload logger to pass requests to components that process them asynchronously. The results can be passed onwards to alterting systems.

![Example architecture](analytics.png)
