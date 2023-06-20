# Resources

For Kubernetes usage we provide a set of custom resources for interacting with Seldon.

 * [SeldonRuntime](./seldonruntime/index.md) - for installing Seldon in a particular namespace.
 * [Servers](./server/index.md) - for deploying sets of replicas of core inference servers (MLServer or Triton). 
 * [Models](./model/index.md) - for deploying single machine learning models, custom transformation logic, drift detectors, outliers detectors and explainers.
 * [Experiments](./experiment/index.md) - for testing new versions of models.
 * [Pipelines](./pipeline/index.md) - for connecting together flows of data between models.

## Advanced Customization

SeldonConfig and ServerConfig define the core installation configuration and machine learning inference server configuration for Seldon. Normally, you would not need to customize these but this may be required for your particular custom installation within your organisation.

 * [ServerConfigs](./serverconfig/index.md) - for defining new types of inference server that can be reference by a Server resource.
 * [SeldonConfig](./seldonconfig/index.md) - for defining how seldon is installed


```{toctree}
:maxdepth: 1
:hidden:

seldonruntime/index.md
model/index.md
experiment/index.md
pipeline/index.md
server/index.md
serverconfig/index.md
seldonconfig/index.md
```

