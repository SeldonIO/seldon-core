# OpenVINO Example with Imagenet

This example shows how an [Intel OpenVINO](https://software.intel.com/en-us/openvino-toolkit) optimized model can be served using Seldon Core. In this case we illustrate using a pre-trained Squeezenet model.

The [notebook](openvino-squeezenet.ipynb) provides a step by step process to:

  * Download the pre-trained model
  * Create your own version of OpenVINO Docker image
  * Run the OpenVINO inference server inside Seldon Core