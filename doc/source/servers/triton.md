# Triton Inference Server

If you have a model that can be run on [NVIDIA Triton Inference Server](https://github.com/triton-inference-server/server) you can use Seldon's Prepacked Triton Server.

Triton has multiple supported backends including support for TensorRT, Tensorflow, PyTorch and ONNX models. For further details see the [Triton supported backends documentation](https://docs.nvidia.com/deeplearning/triton-inference-server/master-user-guide/docs/model_repository.html#section-framework-model-definition).

## Example

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: triton
spec:
  protocol: kfserving
  predictors:
  - graph:
      implementation: TRITON_SERVER
      modelUri: gs://seldon-models/trtis/simple-model
      name: simple
    name: simple
    replicas: 1
```

See more deployment examples in [triton examples](../examples/triton_examples.html) and [protocol examples](../examples/protocol_examples.html).

See also:
- [Tensorflow MNIST - e2e example with MinIO](../examples/triton_mnist_e2e.html)
- [GPT2 Model - pretrained with Azure](../examples/triton_gpt2_example_azure.html)
- [GPT2 Model - pretrained with MinIO](../examples/triton_gpt2_example_azure.html)
