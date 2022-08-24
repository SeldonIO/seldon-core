# Experiment

An Experiment defines a traffic split between Models or Pipelines. This allows new versions of models and pipelines to be tested.

An experiment spec has three sections:

 * `candidates` (required) : a set of candidate models to split traffic.
 * `defaultModel` (optional) : an existing candidate who endpoint should be modified to split traffic as defined by the candidates.
    * Each candidate has a traffic weight. The percentage of traffic will be this weight divided by the sum of traffic weights.
 * `mirror` (optional) : a single model to mirror traffic to the candidates. Responses from this model will not be returned to the caller.

An example experiment with a `defaultModel` is shown below:

```{literalinclude} ../../../../../../samples/experiments/ab-default-model.yaml 
:language: yaml
```

This defines a split of 50% traffic between two models `iris` and `iris2`. In this case we want to expose this traffic split on the existing endpoint created for the `iris` model. This allows us to test new versions of models (in this case `iris2`) on an existing endpoint (in this case `iris`). The `defaultModel` key defines the model whose endpoint we want to change. The experiment will become active when both underplying models are in Ready status.

An experiment over two separate models which exposes a new API endpoint is shown below:

```{literalinclude} ../../../../../../samples/experiments/ab.yaml 
:language: yaml
```

To call the endpoint add the header `seldon-model: <experiment-name>.experiment` in this case: `seldon-model: experiment-iris.experiment`. For example with curl:

```bash
curl http://${MESH_IP}/v2/models/experiment-iris/infer \
   -H "Content-Type: application/json" \
   -H "seldon-model: experiment-iris.experiment" \
   -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

For examples see the [local experiments notebook](../../../examples/local-experiments.md).

## Pipeline Experiments

Running an experiment between some pipelines is very similar. The difference is `resourceType: pipeline` needs to be defined and in this case the candidates or mirrors will refer to pipelines. An example is shown below:

```{literalinclude} ../../../../../../samples/experiments/addmul10.yaml 
:language: yaml
```
For an example see the [local experiments notebook](../../../examples/local-experiments.md).

## Sticky Sessions

To allow cohorts to get consistent views in an experiment each inference request passes back a response header `x-seldon-route` which can be passed in future requests to an experiment to bypass the random traffic splits and get a prediction from the sequence of models and pipelines used in the initial request.

Note: you must pass the normal `seldon-model` header along with the `x-seldon-route` header.

This is illustrated in the [local experiments notebook](../../../examples/local-experiments.md).

Caveats:

  * Note the models used will be the same but not necessarily the same replica instances. This means at present this will not work for stateful models that need to go to the same model replica instance.

## Service Meshes

As an alternative you can choose to run experiments at the service mesh level if you use one of the popular service meshes that allow header based routing in traffic splits. For further discussion see [here](../../service-meshes/index.md).
