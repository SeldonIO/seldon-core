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


## Implementation
The ```EpsilonGreedy``` class implements the epsilon-greedy router.

### Initialization
Required arguments:
* **n_branches**: non-negative integer specifying the number of branches or models to route requests between
* **epsilon**: float in [0,1], defaults to *0.1*

Optional arguments:
* *best_branch*: non-negative integer specifying the initial best performing model. If not provided, selects a random branch
* *verbose*: boolean specifying output verbosity
* *seed*: non-negative integer used to seed the random number generator, intended to use for testing and reproducibility

### Route
Returns the current best branch with probability *1-e*, otherwise returns a random branch with probability *e*.

### Send feedback
Required arguments:
* **features**: array of features for a batch of data samples
* **routing**: non-negative integer specifying the branch for which feedback is sent
* **reward**: float in [0,1] specifying the total total reward for the batch of data samples

The reward is interpreted as the proportion of successes in the batch of data samples. The helper function *n_success_failures* calculates the number of successes and failures given the batch of data samples and the reward.

## Test the source code
A basic test suite is provided in ```test_EpsilonGreedy.py```. Run with ```pytest```.

## Wrap using s2i

```bash
make build
```

## Test the wrapped image

To test the generated docker image using the Seldon Core [internal API](https://github.com/SeldonIO/seldon-core/blob/master/docs/reference/internal-api.md), run it under docker:

```bash
docker run --rm -p 5000:5000 -e PREDICTIVE_UNIT_PARAMETERS='[{"name": "n_branches","value": "3","type": "INT"},{"name": "epsilon","value": "0.3","type": "FLOAT"},{"name": "verbose","value": "1","type": "BOOL"}]' -e PREDICTIVE_UNIT_ID='eg' seldonio/mab_epsilon_greedy:1.3
```
Note that to expose both the ```/route``` and ```/send-feedback``` endpoints we need to provide both ```PREDICTIVE_UNIT_PARAMETERS``` and ```PREDICTIVE_UNIT_ID``` environment variables.

Send a data request:

```bash
data='{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}'
curl -d "json=${data}" http://0.0.0.0:5000/route
```

Send a feedback request:
```bash
data='{"request":{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},"response":{"meta":{"routing":{"eg":2}},"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}},"reward":1}'
curl -d "json=${data}" http://0.0.0.0:5000/send-feedback
```

## Running on Seldon
An end-to-end example running an epsilon-greedy router on GCP to route traffic to 3 models in parallel is available [here](
https://github.com/SeldonIO/seldon-core/blob/master/notebooks/epsilon_greedy_gcp.ipynb) and a Kubeflow integrated example available [here](https://github.com/kubeflow/example-seldon).
