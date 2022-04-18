# Experiment

An Experiment defines a traffic split between Models. This allows new versions of models to be tested and experiments between models to be run.

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