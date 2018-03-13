# Wrapping Your Model

To allow your component (model, router etc) to be managed by seldon-core it needs

 1. To be built into a Docker container
 1. To expose the approripiate [service APIs over REST or gRPC](../reference/internal-api.md).

To wrap your model follow the instructions for your chosen language or toolkit.

## Python

Python based models, including [TensorFlow](https://www.tensorflow.org/), [Keras](https://keras.io/), [pyTorch](http://pytorch.org/), [StatsModels](http://www.statsmodels.org/stable/index.html), [XGBoost](https://github.com/dmlc/xgboost) and [Scikit-learn](http://scikit-learn.org/stable/) based models.

You can use either:

   * [Source-to-image (s2i) tool](./python.md)
   * [Seldon Docker wrapper application](./python-docker.md)

## H2O

   * [H2O models](./h2o.md)

## Future

Future languages:

 * R based models
 * Java based models
 * Go based models
