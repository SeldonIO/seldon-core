# Sequence-to-Sequence LSTM (seq2seq-LSTM) Outlier Detector

## Description

[Anomaly or outlier detection](https://en.wikipedia.org/wiki/Anomaly_detection) has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. 

The implemented seq2seq outlier detector aims to predict anomalies in a sequence of input features. The model can be trained in an unsupervised or semi-supervised way, which is helpful since labeled training data is often scarce. The outlier detector predicts whether the input features represent normal behaviour or not, dependent on a threshold level set by the user.

## Implementation

The architecture of the seq2seq model is defined in ```model.py``` and it is trained by running the ```train.py``` script. The ```OutlierSeq2SeqLSTM``` class loads a pre-trained model and makes predictions on new data.

A detailed explanation of the implementation and usage of the seq2seq model as an outlier detector can be found in the [seq2seq documentation](./doc.md).

## Running on Seldon

An end-to-end example running a seq2seq outlier detector on GCP or Minikube using Seldon to identify anomalies in ECGs is available [here](./seq2seq_lstm.ipynb).

Docker images to use the generic Mahalanobis outlier detector as a model or transformer can be found on Docker Hub:
* [seldonio/outlier-s2s-lstm-model](https://hub.docker.com/r/seldonio/outlier-s2s-lstm-model)
* [seldonio/outlier-s2s-lstm-transformer](https://hub.docker.com/r/seldonio/outlier-s2s-lstm-transformer)

A model docker image specific for the demo is also available:
* [seldonio/outlier-s2s-lstm-model-demo](https://hub.docker.com/r/seldonio/outlier-s2s-lstm-model-demo)