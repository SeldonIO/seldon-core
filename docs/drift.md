# Drift Detection

Drift detection models are treated as any other Model. You can run any saved
[Alibi-Detect](https://github.com/SeldonIO/alibi-detect) drift detection model by
adding the requirement `alibi-detect`.

An example drift detection model from the CIFAR10 image classification example is shown below:

```yaml
# samples/models/cifar10-drift-detect.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: cifar10-drift
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/cifar10/drift-detector"
  requirements:
    - mlserver
    - alibi-detect
```

Usually you would run these models in an asynchronous part of a Pipeline, i.e. they are not
connected to the output of the Pipeline which defines the synchronous path. For example, the
CIFAR-10 image detection example uses a pipeline as shown below:

```yaml
# samples/pipelines/cifar10.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: cifar10-production
spec:
  steps:
    - name: cifar10
    - name: cifar10-outlier
    - name: cifar10-drift
      batch:
        size: 20
  output:
    steps:
    - cifar10
    - cifar10-outlier.outputs.is_outlier
```

Note how the `cifar10-drift` model is not part of the path to the outputs. Drift alerts can be
read from the Kafka topic of the model.

## Examples

* [CIFAR10 image classification with drift detector](examples/cifar10.md)
* [Tabular income classification model with drift detector](examples/income.md)
