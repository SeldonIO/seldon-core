# Mahalanobis Online Outlier Detector

## Description

[Anomaly or outlier detection](https://en.wikipedia.org/wiki/Anomaly_detection) has many applications, ranging from preventing credit card fraud to detecting computer network intrusions.

The Mahalanobis online outlier detector aims to predict anomalies in tabular data. The algorithm calculates an outlier score, which is a measure of distance from the center of the features distribution ([Mahalanobis distance](https://en.wikipedia.org/wiki/Mahalanobis_distance)). If this outlier score is higher than a user-defined threshold, the observation is flagged as an outlier. The algorithm is online, which means that it starts without knowledge about the distribution of the features and learns as requests arrive. Consequently you should expect the output to be bad at the start and to improve over time.

## Implementation

The algorithm is implemented in the ```OutlierMahalanobis``` class and a detailed explanation of the implementation and usage of the algorithm to spot anomalies can be found in the [outlier_mahalanobis_doc](./outlier_mahalanobis_doc.ipynb) notebook.

## Running on Seldon

An end-to-end example running a Mahalanobis outlier detector on GCP or Minikube using Seldon to identify computer network intrusions is available [here](./outlier_mahalanobis.ipynb).