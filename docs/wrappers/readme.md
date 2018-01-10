# Wrapping Your Model

Seldon-core deploys dockerized versions of your models, which means that  you will need to wrap your model into a docker image. Once the docker image of your model has been built, Seldon-core will run it as a container in a kubernetes cluster and access your model via seldon API server.

Seldon-core inherits model agnosticity from docker, which means you can deploy  models written in any programming language. The only requirement is that the input and output messages sent to the docker container where your model is deployed respects the internal [Seldon Microservice API](../reference/internal-api.md).

# Seldon Wrappers
Seldon provides some tools for popular machine learning toolkits and languages to help wrapping your model. Currently Seldon Core provides the following wrappers:

* [python models](./python.md)
  * including keras, tensorflow and sklearn models.
* [H2O models](./h2o.md)
