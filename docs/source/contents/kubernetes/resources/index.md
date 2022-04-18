# Resources

For Kubernetes usage we provide a set of custom resources for interacting with Seldon.

 * [Models](./model/index.md) - for deploying single machine learning models, custom transformation logic, drift detectors and outliers detectors.
 * [Experiments](./experiment/index.md) - for testing new versions of models
 * [Pipelines](./pipeline/index.md) - for connecting together flows of data between models 
 * [Explainers](./explainer/index.md) - model explanations. (coming soon)

Advanced usage:

 * [Servers](./server/index.md) - for deploying sets of replicas of core inference servers (MLServer or Triton).
 * [ServerConfigs](./serverconfig/index.md) - for defining new types of inference server that can be reference by a Server resource.


```{toctree}
:maxdepth: 1
:hidden:

model/index.md
experiment/index.md
pipeline/index.md
explainer/index.md
server/index.md
serverconfig/index.md
```

