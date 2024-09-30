# Explainers

Explainers are Model resources with some extra settings. They allow a range of explainers
from the Alibi-Explain library to be run on MLServer.

An example Anchors explainer definitions is shown below.

```yaml
# samples/models/income-explainer.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-explainer
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.5.0/income-sklearn/anchor-explainer"
  explainer:
    type: anchor_tabular
    modelRef: income
```

The key additions are:

* `type`: This must be one of the
[supported Alibi Explainer types](https://github.com/SeldonIO/MLServer/blob/191ee44297712192fed882afe0797d6a2732965e/runtimes/alibi-explain/mlserver_alibi_explain/alibi_dependency_reference.py#L15-L19)
supported by the Alibi Explain runtime in MLServer.
* `modelRef`: The model name for black box explainers.
* `pipelineRef`: The pipeline name for black box explainers.

Only one of modelRef and pipelineRef is allowed.

## Pipeline Explanations

Blackbox explainers can explain a Pipeline as well as a model. An example from the [Huggingface sentiment demo](../examples/speech-to-sentiment.md) is show below.

```yaml
# samples/models/hf-sentiment-explainer.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: sentiment-explainer
spec:
  storageUri: "gs://seldon-models/scv2/examples/huggingface/speech-sentiment/explainer"
  explainer:
    type: anchor_text
    pipelineRef: sentiment-explain
```

## Examples
* [Tabular income classification model with Anchor Tabular black box model explainer](examples/income.md)
* [Huggingface Sentiment model with Anchor Text black box pipeline explainer](examples/speech-to-sentiment.md)
* [Anchor Text movies sentiment explainer](examples/explainer-examples.md)
