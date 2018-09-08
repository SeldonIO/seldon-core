# Wrapping Your Model

To allow your component (model, router etc) to be managed by seldon-core it needs

1.  To be built into a Docker container
1.  To expose the approripiate [service APIs over REST or gRPC](../reference/internal-api.md).

To wrap your model follow the instructions for your chosen language or toolkit.

## Python

Python based models, including [TensorFlow](https://www.tensorflow.org/), [Keras](https://keras.io/), [pyTorch](http://pytorch.org/), [StatsModels](http://www.statsmodels.org/stable/index.html), [XGBoost](https://github.com/dmlc/xgboost) and [Scikit-learn](http://scikit-learn.org/stable/) based models.

You can use either:

- [Source-to-image (s2i) tool](./python.md)
- [Seldon Docker wrapper application](./python-docker.md)

## R

- [R models can be wrapped using source-to-image](r.md)

## Java

Java based models including, [H2O](https://www.h2o.ai/), [Deep Learning 4J](https://deeplearning4j.org/), Spark (standalone exported models).

- [Java models wrapped using source-to-image](java.md)

## H2O

H2O models can be wrapped either from Java or Python.

- [Java models wrapped using source-to-image](java.md)
- [H2O models saved and called from python](./h2o.md)

## NodeJS

- [Javascript models can be wrapped using source-to-image](nodejs.md)
