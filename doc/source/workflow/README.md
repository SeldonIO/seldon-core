
# Seldon Core Workflow

There are 3 steps to using seldon-core.

 1. Install seldon-core onto a Kubernetes cluster
 1. Wrap your components (usually runtime model servers) as Docker containers that respect the internal Seldon microservice API.
 1. Define your runtime service graph as a SeldonDeployment resource and deploy your model and serve predictions

![steps](./steps.png)

## Install Seldon Core

To install seldon-core follow the [installation guide](install.md).

## Wrap Your Model

The components you want to run in production need to be wrapped as Docker containers that respect the [Seldon microservice API](../reference/apis/internal-api.md). You can create models that serve predictions, routers that decide on where requests go, such as A-B Tests, Combiners that combine responses and transformers that provide generic components that can transform requests and/or responses.

To allow users to easily wrap machine learning components built using different languages and toolkits we provide wrappers that allow you easily to build a docker container from your code that can be run inside seldon-core. Our current recommended tool is RedHat's Source-to-Image. More detail can be found in [Wrapping your models docs](../wrappers/README.md).

## Define Runtime Service Graph

To run your machine learning graph on Kubernetes you need to define how the components you created in the last step fit together to represent a service graph. This is defined inside a `SeldonDeployment` Kubernetes Custom resource. A [guide to constructing this inference graph is provided](../graph/inference-graph.md).

![graph](./graph.png)

## Deploy and Serve Predictions

You can use ```kubectl``` to deploy your ML service like any other Kubernetes resource. This is discussed [here](deploying.md). Once deployed ypu can get predictions by [calling the exposed API](serving.md).

## Next Steps

Run a [notebook](../examples/helm_examples.html)  using Helm that illustrates using our Helm charts for launching various types of inference graphs.



