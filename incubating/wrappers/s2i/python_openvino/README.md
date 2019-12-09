# Seldon base image with python and OpenVINO inference engine

## Overview

Seldon prediction base component with [OpenVINO toolkit](https://software.intel.com/en-us/openvino-toolkit) 
makes it easy to implement inference operation with performance boost.

OpenVINO inference engine together with model optimizer makes it possible to achieve faster execution.

Use [model optimizer](https://software.intel.com/en-us/articles/OpenVINO-ModelOptimizer) to convert trained models from
frameworks like TensorFlow, MXNET, Caffe, Kaldi or ONNX to Intermediate Representation format.

It can be used more efficiently to execute inference operations using 
[inference engine](https://software.intel.com/en-us/articles/OpenVINO-InferEngine).

It will take advantage of all the CPU features to reduce the inference latency and gain extra throughput.

Current version of OpenVINO supports also 
[low precision models](https://www.intel.ai/introducing-int8-quantization-for-fast-cpu-inference-using-openvino),
which improve the performance even more. At the same time
the accuracy impact is minimal.


## Building
```bash

make build
```
## Usage

This base image can be used to Seldon components exactly the same way like with standard Seldon base images.
Use s2i tool like documented [here](https://github.com/SeldonIO/seldon-core/blob/master/docs/wrappers/python.md).
An example is presented below:

```bash
s2i build . seldonio/seldon-core-s2i-openvino:0.2 {target_component_image_name}
```

## Examples

[Models ensemble with OpenVINO](../../../examples/models/openvino_imagenet_ensemble)

## References

[OpenVINO toolkit](https://software.intel.com/en-us/openvino-toolkit)

[OpenVINO API docs](https://software.intel.com/en-us/articles/OpenVINO-InferEngine#inpage-nav-9)

[Seldon pipeline example](../../../examples/models/openvino_imagenet_ensemble)


## Notes

This Seldon base image contains, beside OpenVINO inference execution engine python API also several other useful components.
- Intel optimized python version
- Intel optimized OpenCV package
- Intel optimized TensorFlow with MKL engine
- Configured conda package manager

OpenVINO and TensorFlow in this docker image, employs [MKL-DNN](https://github.com/intel/mkl-dnn) library
with [OpenMP](https://www.openmp.org/) threading control.
Make sure you configure optimal values for MKL related environment variables in the containers.
Recommendations are listed below:

`KMP_AFFINITY`=granularity=fine,verbose,compact,1,0

`KMP_BLOCKTIME`=1

`OMP_NUM_THREADS`={number of physical CPU cores to allocate}