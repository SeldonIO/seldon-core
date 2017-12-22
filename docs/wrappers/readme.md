# Seldon wrappers guide

Seldon-core deploys dockerized versions of your models, so in order to use Seldon-core you will need to wrap your model into a docker image. Once the docker image of your model is builded, Seldon-core will run it as a container in a kubernetes cluster and access your model via seldon API server.

Seldon-core inherits model  agnosticity from docker, which means you can deploy any model written using any tool. All it is required is that the input and output messages sent to the docker container where your model is deployed respect [Seldon API]().

To ensure that, seldon-core uses wrappers to dockerize your model. Currently, Seldon-core inculde builded-in wrappers for python models and H2O models. You can find how-to guides for model wrapping in the following links:

* [python models](./python.md), including keras, tensorflow and sklearn models.
* [H2O models](./h2o.md)