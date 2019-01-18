# Outlier Detection in Seldon Core

## Description

[Anomaly or outlier detection](https://en.wikipedia.org/wiki/Anomaly_detection) has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. Seldon Core provides a number of outlier detectors suitable for different use cases. The detectors can be run as a model which is one of the pre-defined types of [predictive units](../../docs/reference/seldon-deployment.md#proto-buffer-definition) in Seldon Core. It is a microservice that makes predictions and can receive feedback rewards. The REST and gRPC internal APIs that the model components must conform to are covered in the [internal API](../../docs/reference/internal-api.md#model) reference.


## Implementations

The following types of outlier detectors are implemented and showcased with demos on Seldon Core:
* [Sequence-to-Sequence LSTM](./seq2seq-lstm)
* [Variational Auto-Encoder](./vae)
* [Isolation Forest](./isolation-forest)
* [Mahalanobis Distance](./mahalanobis)

The Sequence-to-Sequence LSTM algorithm can be used to detect outliers in time series data, while the other algorithms spot anomalies in tabular data. The Mahalanobis detector works online and does not need to be trained first. The other algorithms are ideally trained on a batch of normal data or data with a low fraction of outliers.

## Language specific templates

A reference template for custom model components written in several languages are available:
* [Python](../../wrappers/s2i/python/test/model-template-app/MyModel.py)
* [R](../../wrappers/s2i/R/test/model-template-app/MyModel.R)

Additionally, the [wrappers](../../wrappers/s2i) provide guidelines for implementing the model component in other languages.