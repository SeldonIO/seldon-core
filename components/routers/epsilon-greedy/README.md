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

**NB:** The reward is interpreted as the proportion of successes in the batch of data samples. Thus this implementation inherently assumes binary rewards for each sample in the batch. The helper function *n_success_failures* calculates the number of successes and failures given the batch of data samples and the reward.

This means that our version of the epsilon-greedy router solves the **Bernoulli** bandit.

## Test the source code
A basic test suite is provided in ```test_EpsilonGreedy.py```. Run with ```pytest```.

## Case Study
You can find a case study comparing epsilon-greedy routing and Thompson sampling used as routers for models predicting credit card default [here](../case_study/credit_card_default.ipynb).

## Pre-wrapped image
The latest version of the Epsilon Greedy Router available from Docker Hub is [```seldonio/mab_epsilon_greedy:1.3```](https://hub.docker.com/r/seldonio/mab_epsilon_greedy).

## Wrap using s2i
### Persistence
For routers like multi-armed bandits it can be important to save the state after some learning has been done to avoid cold starts when re-deploying an inference graph. This can be achieved by setting ```PERSISTENCE=1``` in the ```.s2i/environment``` file before wrapping the source code. This will use redis to periodically save state of the component on the Seldon Core cluster.

### Wrap
```bash
make build
```

## Test the wrapped image

See the [example notebook](egreedy.ipynb)


