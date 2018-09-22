# NVIDIA Inference Server Proxy

The NVIDIA Inference Server Proxy ('''seldonio/nvidia-inference-server-proxy''') provides a proxy to forward Seldon prediction requests to a running [NVIDIA Inference Server](https://docs.nvidia.com/deeplearning/sdk/inference-user-guide/index.html).

The intiial release assumes you have a running inference server. Examples to get a running inference server include:
  * [Kubeflow Integration](https://github.com/kubeflow/kubeflow/tree/master/kubeflow/nvidia-inference-server)


In future releases, we plan to package a running inference server next to the proxy to allow better management of resources in one place via Seldon.

Examples:

 * [MNIST with Nvidia Inference Server](../examples/models.
