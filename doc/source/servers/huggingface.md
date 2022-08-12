# HuggingFace Server

Thanks to our collaboration with the HuggingFace team you can now easily deploy your models from the [HuggingFace Hub](https://huggingface.co/models) with Seldon Core.

We also support the high performance optimizations provided by the [Transformer Optimum framework](https://huggingface.co/docs/optimum/index).

## Pipeline parameters

The parameters that are available for you to configure include:

| Name | Description |
| ---- | ----------- |
| `task` | The transformer pipeline task |
| `pretrained_model` | The name of the pretrained model in the Hub |
| `pretrained_tokenizer` | Transformer name in Hub if different to the one provided with model |
| `optimum_model` | Boolean to enable loading model with Optimum framework |

## Simple Example

You can deploy a HuggingFace model by providing parameters to your [pipeline](https://huggingface.co/docs/transformers/main_classes/pipelines).

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: gpt2-model
spec:
  protocol: v2
  predictors:
  - graph:
      name: transformer
      implementation: HUGGINGFACE_SERVER
      parameters:
      - name: task
        type: STRING
        value: text-generation
      - name: pretrained_model
        type: STRING
        value: distilgpt2
    name: default
    replicas: 1
```

## Quantized & Optimized Models with Optimum

You can deploy a HuggingFace model loaded using the Optimum library by using the `optimum_model` parameter

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: gpt2-model
spec:
  protocol: v2
  predictors:
  - graph:
      name: transformer
      implementation: HUGGINGFACE_SERVER
      parameters:
      - name: task
        type: STRING
        value: text-generation
      - name: pretrained_model
        type: STRING
        value: distilgpt2
      - name: optimum_model
        type: BOOL
        value: true
    name: default
    replicas: 1
```

