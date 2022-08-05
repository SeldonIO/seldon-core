# Source to Image (s2i)

[Source to image](https://github.com/openshift/source-to-image) is a RedHat supported tool to create docker images from source code. We provide builder images to allow you to easily wrap your data science models so they can be managed by seldon-core.

The general work flow is:

 1. [Download and install s2i](https://github.com/openshift/source-to-image#installation)

 1. Choose the builder image that is most appropriate for your code and get usage instructions, for example:

    ```bash
    s2i usage seldonio/seldon-core-s2i-python3
    ```

 1. Create a source code repo in the form acceptable for the builder image and build your docker container from it. Below we show an example using our seldon-core git repo which has some template examples for python models.

    ```bash
    s2i build https://github.com/seldonio/seldon-core.git \
        --context-dir=wrappers/s2i/python/test/model-template-app seldonio/seldon-core-s2i-python3:1.14.1 \
        seldon-core-template-model
    ```

At present we have s2i builder images for

 * [Python (Python3)](../python/README.md) : use this for Tensorflow, Keras, PyTorch or sklearn models.
 * [R](../R/README.md)
 * [Java](../java/README.md)
 * [NodeJS](../nodejs/README.md)

