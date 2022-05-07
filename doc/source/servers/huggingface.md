# HuggingFace Server

With our recent collaboration with the HuggingFace team you can now easily deploy your models from the HuggingFace Hub.

We also support the high performance optimizations provided by the Transformer Optimum server.

## Simple Example

You can deploy a HuggingFace model by providing parameters to your pipeline

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

The parameters that are available for you to configure include:

| Name | Description |
| ---- | ----------- |
| `task` | The transformer pipeline task |
| `pretrained_model` | The name of the pretrained model in the Hub |
| `pretrained_tokenizer` | Transformer name in Hub if different to the one provided with model |
| `optimized` | Whether to attempt to load with Optimum framework |


