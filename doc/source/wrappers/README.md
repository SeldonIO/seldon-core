# Wrapping Your Model

To allow your component (model, router etc.) to be managed by seldon-core it needs

1.  To be built into a Docker container
1.  To expose the appropriate [service microservice APIs over REST or gRPC](../reference/apis/internal-api.md).

To wrap your model follow the instructions for your chosen language or toolkit.

To test a wrapped components you can use one of our [testing scripts](../workflow/api-testing.md).

## Python

Python based models, including [TensorFlow](https://www.tensorflow.org/), [Keras](https://keras.io/), [pyTorch](http://pytorch.org/), [StatsModels](http://www.statsmodels.org/stable/index.html), [XGBoost](https://github.com/dmlc/xgboost) and [Scikit-learn](http://scikit-learn.org/stable/) based models.

- [Python model wrapping](../python/index.html)

## R

- [R models wrapped using source-to-image](../R/README.md)

## Java

Java based models including, [H2O](https://www.h2o.ai/), [Deep Learning 4J](https://deeplearning4j.org/), Spark (standalone exported models).

- [Java models wrapped using source-to-image](../java/README.md)

## NodeJS

- [Javascript models wrapped using source-to-image](../nodejs/README.md)


## Go (Alpha)

- [Example Go integration](https://github.com/SeldonIO/seldon-core/tree/master/examples/wrappers/go)

