# Seldon base image with python and OpenVINO inference engine

## Building
```bash

cp ../python/s2i .
docker build -f Dockerfile_openvino_base --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy \
-t seldon_openvino_base:latest .
```
## Usage

This base image can be used to Seldon components exactly the same way like with standard Seldon base images.
Use s2i tool like documented [here](https://github.com/SeldonIO/seldon-core/blob/master/docs/wrappers/python.md).
An example is presented below:

```bash
s2i build . seldon_openvino_base:latest {component_image_name}
```

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

In case you would use this compoment to run inference operations using TensorFlow with MKL, make sure you configure 
also the following environment variables:

`KMP_AFFINITY`=granularity=fine,verbose,compact,1,0

`KMP_BLOCKTIME`=1

`OMP_NUM_THREADS`={number of CPU cores}