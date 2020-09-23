# Triton Inference Server

If you have a model that can be run on [NVIDIA Triton Inference Server](https://github.com/triton-inference-server/server) you can use Seldon's Prepacked Triton Server.

Triton has multiple supported backends including support for TensorRT, Tensorflow, PyTorch and ONNX models. For further details see the [Triton supported backends documentation](https://docs.nvidia.com/deeplearning/triton-inference-server/master-user-guide/docs/model_repository.html#section-framework-model-definition).

## Example

```
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

Try out a [worked notebook](../examples/protocol_examples.html)
