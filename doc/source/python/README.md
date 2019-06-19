# Python Components

To create a seldon Component in Python:

 1. Install the [Seldon Core Python Module](python_module.md)
 1. [Create your Python class](python_component.md) for your model or component
 1. [Wrap the component using S2I](python_wrapping_s2i.md) or [Docker](python_wrapping_docker.md).

You can find various examples in our [example notebooks](../examples/notebooks.html).

It also possible to mount a model into an existing image. This can be seen in the [seldon kubeflow examples](https://github.com/kubeflow/example-seldon)