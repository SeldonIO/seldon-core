# Wrapping Your Model

Seldon-core deploys dockerized versions of your models, meaning that your models need to be turned into docker images, and expose an API that conforms to Seldon. We call this process wrapping a model. Once the docker image of your model has been built, Seldon-core will run it as a container inside a kubernetes cluster and all requests must be sent to Seldon's unified API before being routed to your model.

Seldon-core is model agnostic: you can deploy models written in any programming language. The only requirement is that your dockerized model conforms to the internal [Seldon Microservice API](../reference/internal-api.md).

# Seldon Wrappers
Seldon provides tools to automatically wrap models from popular machine learning toolkits and languages, called wrappers. Currently Seldon Core provides the following wrappers:

* [python models](./python.md)
  * including keras, tensorflow and sklearn models.
* [H2O models](./h2o.md)

These wrappers can be used by persons without expertise in docker or microservices.
