# Isolation Forest Outlier Detector

## Description

[Anomaly or outlier detection](https://en.wikipedia.org/wiki/Anomaly_detection) has many applications, ranging from preventing credit card fraud to detecting computer network intrusions. The implemented [Isolation Forest](https://scikit-learn.org/stable/modules/generated/sklearn.ensemble.IsolationForest.html) outlier detector aims to predict anomalies in tabular data. The anomaly detector predicts whether the input features represent normal behaviour or not, dependent on a threshold level set by the user.

## Implementation

The Isolation Forest is trained by running the ```train.py``` script. The ```OutlierIsolationForest``` class loads a pre-trained model and makes predictions on new data.

A detailed explanation of the implementation and usage of Isolation Forests as outlier detectors can be found in the [isolation_forest_doc](./isolation_forest_doc.ipynb) notebook.

## Running on Seldon

An end-to-end example running an Isolation Forest outlier detector on GCP or Minikube using Seldon to identify computer network intrusions is available [here](./isolation_forest.ipynb).