# Variational Auto-Encoder (VAE) Outlier Detector

## Description

[Anomaly or outlier detection](https://en.wikipedia.org/wiki/Anomaly_detection) has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. The implemented VAE outlier detector aims to predict anomalies in tabular data. The VAE model can be trained in an unsupervised or semi-supervised way, which is helpful since labeled training data is often scarce. The outlier detector predicts whether the input features represent normal behaviour or not, dependent on a threshold level set by the user.

## Implementation

The architecture of the VAE is defined in ```model.py``` and the model is trained by running the ```train.py``` script. The ```OutlierVAE``` class loads a pre-trained model and makes predictions on new data.

A detailed explanation of the implementation and usage of the Variational Auto-Encoder as an outlier detector can be found in the [VAE documentation](./doc.md). 

## Running on Seldon

An end-to-end example running a VAE outlier detector on GCP or Minikube using Seldon to identify computer network intrusions is available [here](./outlier_vae.ipynb).

Docker images to use the generic VAE outlier detector as a model or transformer can be found on Docker Hub:
* [seldonio/outlier-vae-model](https://hub.docker.com/r/seldonio/outlier-vae-model)
* [seldonio/outlier-vae-transformer](https://hub.docker.com/r/seldonio/outlier-vae-transformer)

A model docker image specific for the demo is also available:
* [seldonio/outlier-vae-model-demo](https://hub.docker.com/r/seldonio/outlier-vae-model-demo)