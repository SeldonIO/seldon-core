# Outlier Detection in Seldon Core

## Description

[Anomaly or outlier detection](https://en.wikipedia.org/wiki/Anomaly_detection) has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. Seldon Core provides a number of outlier detectors suitable for different use cases. The detectors can be run as models or transformers which are part of the pre-defined types of [predictive units](../../docs/reference/seldon-deployment.md#proto-buffer-definition) in Seldon Core. Models are microservices that make predictions and can receive feedback rewards while the input transformers add the anomaly predictions to the metadata of the underlying model. The REST and gRPC internal APIs that the model and transformer components must conform to are covered in the [internal API](../../docs/reference/internal-api.md) reference.

## Implementations

The following types of outlier detectors are implemented and showcased with demos on Seldon Core:
* [Sequence-to-Sequence LSTM](./seq2seq-lstm)
* [Variational Auto-Encoder](./vae)
* [Isolation Forest](./isolation-forest)
* [Mahalanobis Distance](./mahalanobis)

The Sequence-to-Sequence LSTM algorithm can be used to detect outliers in time series data, while the other algorithms spot anomalies in tabular data. The Mahalanobis detector works online and does not need to be trained first. The other algorithms are ideally trained on a batch of normal data or data with a low fraction of outliers.

## Implementing custom outlier detectors

An outlier detection component can be implemented either as a model or input transformer component. If the component is defined as a model, a ```predict``` method needs to be implemented to return the detected anomalies. Optionally, a ```send_feedback``` method can return additional information about the performance of the algorithm. When the component is used as a transformer, the anomaly predictions will occur in the ```transform_input``` method which returns the unchanged input features. The anomaly predictions will then be added to the underlying model's metadata via the ```tags``` method. Both models and transformers can make use of custom metrics defined by the ```metrics``` function. 

The required methods to use the outlier detection algorithms as models or transformers are implemented in the Python files with the ```Core``` prefix. The demos contain clear instructions on how to run your component as a model or transformer.

## Language specific templates

Reference templates for custom model and input transformer components written in several languages are available:
* Python
  * [model](../../wrappers/s2i/python/test/model-template-app/MyModel.py)
  * [transformer](../../wrappers/s2i/python/test/transformer-template-app/MyTransformer.py)
* R
  * [model](../../wrappers/s2i/R/test/model-template-app/MyModel.R)
  * [transformer](../../wrappers/s2i/R/test/transformer-template-app/MyTransformer.R)

Additionally, the [wrappers](../../wrappers/s2i) provide guidelines for implementing the model component in other languages.