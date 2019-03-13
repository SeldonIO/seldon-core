# Inference Graph

Seldon Core extends Kubernetes with its own custom resource SeldonDeployment where you can define your runtime inference graph made up of models and other components that Seldon will manage.

A SeldonDeployment is a JSON or YAML file that allows you to define your graph of component images and the resources each of those images will need to run (using a Kubernetes PodTemplateSpec). The parts of a SeldonDeployment are shown below:

![inference-graph](./inf-graph.png)

A minimal example for a single model, this time in YAML, is shown below:
```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  name: test-deployment
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: example
    replicas: 1
```

The key components are:

  * A list of Predictors, each with a specification for the number of replicas.
     * Each defines a graph and its set of deployments. Multiple predictors is useful when you want to split traffic between a main graph and a canary or for other production rollout scenarios.
  * For each predictor a list of componentSpecs. Each componentSpec is a Kubernetes PodTemplateSpec which Seldon will build into a Kubernetes Deployment. Place here the images from your graph and their requirements, e.g. Volumes, ImagePullSecrets, Resources Requests etc.
  * A graph specification that describes how your components are joined together.

To understand the inference graph definition in detail see [here](../reference/apis/crd.md)


