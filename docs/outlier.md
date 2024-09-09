# Outlier Detection

Outlier detection models are treated as any other Model. You can run any saved
[Alibi-Detect](https://github.com/SeldonIO/alibi-detect) outlier detection model
by adding the requirement `alibi-detect`.

An example outlier detection model from the CIFAR10 image classification example is shown below:

```yaml
# samples/models/cifar10-outlier-detect.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: cifar10-outlier
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/cifar10/outlier-detector"
  requirements:
    - mlserver
    - alibi-detect
```

## Examples

* [CIFAR10 image classification with outlier detector](../examples/cifar10.md)
* [Tabular income classification model with outlier detector](../examples/income.md)
