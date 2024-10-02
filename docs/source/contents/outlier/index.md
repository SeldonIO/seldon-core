# Outlier Detection

Outlier detection models are treated as any other Model. You can run any saved [Alibi-Detect](https://github.com/SeldonIO/alibi-detect) outlier detection model by adding the requirement `alibi-detect`.

An example outlier detection model from the CIFAR10 image classification example is shown below:

```{literalinclude} ../../../../samples/models/cifar10-outlier-detect.yaml
:language: yaml
```

## Examples

 * [CIFAR10 image classification with outlier detector](../examples/cifar10.md)
 * [Tabular income classification model with outlier detector](../examples/income.md)
