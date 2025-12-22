---
description: Optimizing the model artefact in Seldon Core 2.
---

## Optimizing the model artefact

The speed at which an ML model can return results given input is based on the model’s architecture, model size, the precision of the model’s weights, and input size. In order to reduce the inherent complexity in the data processing required to execute an inference due to the attributes of a model, it is worth considering: 

- **Model pruning** to reduce parameters that may be unimportant. This can help reduce model size without having a big impact on the quality of the model’s outputs.
- **Quantization** to reduce the computational and memory overheads of running inference by using model weights and activations with lower precision data types.
- **Dimensionality reduction** of inputs to reduce the complexity of computation.
- **Efficient model architectures** such as MobileNet, EfficientNet, or DistilBERT, which are designed for faster inference with minimal accuracy loss.
- **Optimized model formats and runtimes** like ONNX Runtime, TensorRT, or OpenVINO, which leverage hardware-specific acceleration for improved performance.


