# Seldon Documentation

## Quick Start

 - [Quick Start using Minikube](./getting_started/minikube.md)
 - [Jupyter Notebook showing deployment of prebuilt model](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/kubectl_demo_minikube.ipynb)

## Deployment Guide

![API](./deploy.png)

Three steps:

 1. [Wrap your runtime prediction model](./wrappers/readme.md).
 1. [Define your runtime inference graph in a seldon deployment custom resource](./crd/readme.md).
 1. [Deploy the graph](./deploying.md).

## Reference

 - [Prediction API](./reference/prediction.md)
 - [Seldon Deployment Custom Resource](./reference/seldon-deployment.md)
