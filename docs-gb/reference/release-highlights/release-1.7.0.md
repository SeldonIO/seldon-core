# Seldon Core Release 1.7.0

A summary of the main contributions to the [Seldon Core release 1.7.0](https://github.com/SeldonIO/seldon-core/releases/tag/v1.7.0).

## Experimental GPU Accelerated Explainers and Drift Detection

As part of our NVIDIA GTC 2021 Talk this year we have added some new features for utilizing GPUs for explainability and drift detection.

### XGBoost Model with GPU TreeShap Explainer

We have added an example for GPU accelerated TreeShap Explainers which show how explanations can be optimized using NVIDIA GPUs to achieve significant speedups when processing explanations as compared to the CPU version of TreeShap.

The example is available in the [GPU Example Section of the Explainer Seldon Core Notebook](https://docs.seldon.io/projects/seldon-core/en/v1.7.0/examples/explainer_examples.html#Experimental:-XGBoost-Model-with-GPU-TreeShap-Explainer). If you have a GPU accelerated cluster, you can try deploying the model with the respective explainer with the following YAML:

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: incomegpu
spec:
  annotations:
    seldon.io/rest-timeout: "100000"
  predictors:
  - graph:
      children: []
      implementation: XGBOOST_SERVER
      modelUri: gs://seldon-models/xgboost/adult/model_1.0.2
      name: income-model
    explainer:
      type: TreeShap
      modelUri: gs://seldon-models/xgboost/adult/tree_shap_gpu
      containerSpec:
        name: explainer
        image: seldonio/alibiexplainer-gpu:1.7.0-dev
        resources:
          limits:
            nvidia.com/gpu: 1
    name: default
    replicas: 1
```

## Drift Detection with GPU Accelerated Triton Inference Server and Drift Detector

We provide an example of deploying a CIFAR10 image classification model on Triton Inference Server alongside a GPU accelerated drift detector utilizing KNative. The architecture is as shown below:

![Drift Architecture](./drift-gpu.png)

The [example notebook](https://github.com/SeldonIO/seldon-core/blob/master/components/drift-detection/nvidia-triton-cifar10/cifar10_drift.ipynb) illustrates the steps to deploy the model and drift detector and test drift.

## Distributed Persistent State for Multi-Armed Bandits

In production use-cases of multi-armed bandits and generally other type of online-learning models, the concept of distributed state has growingly become a priority. 

In this release we have added an extension to the Multi-Armed Bandit Thomson Sampling example which is implemented using Redis to manage distributed state, which  allows for consistency when increasing the number of replicas, as well as providing the ability to update the state of the Multi-Armed Bandit from a completely separate component/microservice, as opposed to within the same SeldonDeployment.

You can try the distributed Thomson Sampling multiarmed bandit example in the [MAB case study notebook](https://github.com/SeldonIO/seldon-core/blob/master/components/routers/case_study/credit_card_default.ipynb).

## Storage Initializer Customisation on Seldon Deployment

With version 1.7.0 it is now possible to specify image used for the Storage Initializers used with Pre-Packaged Model servers on each Seldon Deployment CR.
```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: custom-sklearn
spec:
  predictors:
  - name: default
    replicas: 1
    graph:
      name: classifier
      implementation: SKLEARN_SERVER
      modelUri: mys3:sklearn/iris
      storageInitializerImage: kfserving/storage-initializer:v0.6.1           # Specify custom image here
      envSecretRefName: seldon-init-container-secret                          # Specify custom secret here
```

## Security Vulnerability Patches

We have updated our base Python images to address CVEs identified, which aligns to the Seldon Core policy. This further strengthens the security of Seldon Core by ensuring that not only the dependencies are updated to address vulnerabilities, but now the containers have been scanned to identify other vulnerabilities.

## Other highlights

* Updated request logging examples to use OpenDistro for Elasticsearch
*
