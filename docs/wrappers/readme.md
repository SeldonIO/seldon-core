# Wrapping Your Model

To allow your component (model, router etc) to be managed by seldon-core it needs

 1. To be built into a Docker container
 1. Expose the approripiate [service APIs over REST or gRPC](../reference/internal-api.md).

To allow developers to more easily wrap their runtime models we provide instructions to use RedHat's Source-to-image tool s2i.

Read [general instructions on using s2i](./s2i.md) and then follow the links below for wrapping instructions for your language/ML tool set of choice:

 * [Python based models](./python.md), including TensorFlow, Keras, pyTorch and sklearn based models

Future languages:

 * R based models
 * Java based models
 * Go based models

Particular ML Toolkits:

   * [H2O models](./h2o.md)

These wrappers can be used by persons without expertise in docker or microservices.
