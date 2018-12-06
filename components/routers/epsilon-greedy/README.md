# Epsilon Greedy Router

## Description

An epsilon-greedy router implements a [multi-armed bandit strategy](https://en.wikipedia.org/wiki/Multi-armed_bandit#Semi-uniform_strategies) in which, when presented with *n* models to make predictions, the currently
best performing model is selected with probability *1-e* while a random model is selected with probability *e*.
This strategy ensures sending traffic to the best performing model most of the time (exploitation) while allowing for
some evaluation of other models (exploration). A typical parameter value could be *e=0.1*, but this will depend on the
desired trade-off between exploration and exploitation.

Note that in this implementation the parameter value *e* is static, but a related strategy called *epsilon-decreasing*
would see the value of *e* decrease as the number of predictions increases, resulting in a highly explorative behaviour
at the start and increasingly exploitative behaviour as time goes on.


## Wrap using s2i

```bash
s2i build . seldonio/seldon-core-s2i-python3 egreedy-router
```

## Smoke Test

Run under docker.

```bash
docker run --rm -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"name": "n_branches","value": "3","type": "INT"},{"name": "epsilon","value": "0.3","type": "FLOAT"},{"name": "verbose","value": "1","type": "BOOL"}]' egreedy-router
```

Send a data request.

```bash
data='{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
curl -d "json=${data}" http://0.0.0.0:5000/route
```

## Running on Seldon
An end-to-end example deploying an epsilon-greedy router to route traffic to 3 models in parallel is available [here](
https://github.com/SeldonIO/seldon-core/blob/master/notebooks/epsilon_greedy_gcp.ipynb).
