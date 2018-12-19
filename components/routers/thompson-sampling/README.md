# Thompson Sampling Router

## Description

[Thompson Sampling](https://en.wikipedia.org/wiki/Thompson_sampling) is a Bayesian method for solving the [multi-armed bandit problem](https://en.wikipedia.org/wiki/Multi-armed_bandit#Semi-uniform_strategies). The main idea is to start with a prior distribution over the unknown reward distributions and use each pull of the arm to update the posterior hyperparameters. Each subsequent pull of the arm is then decided by sampling from the updated posterior distributions over the rewards.

## Implementation
The ```ThompsonSampling``` class implements the Thompson Sampling router.

**NB:** The reward is interpreted as the proportion of successes in the batch of data samples. Thus this implementation inherently assumes binary rewards for each sample in the batch. The helper function *n_success_failures* calculates the number of successes and failures given the batch of data samples and the reward.

This means that our version of the Thompson Sampling router implements the **Beta-Bernoulli** model. The Beta distribution is a conjugate prior to the Bernoulli distribution allowing for a very simple hyperparameter update. We set our prior hyperparameters α=β=1 corresponding to a uniform distribution over the number of arms. For a thorough discussion on the Beta-Bernoulli model refer to Chapter 3 of [A Tutorial on Thompson Sampling](https://arxiv.org/abs/1707.02038).

## Case Study
You can find a case study comparing epsilon-greedy routing and Thompson sampling used as routers for models predicting credit card default [here](../case_study/credit_card_default.ipynb).

## Pre-wrapped image
The latest version of the Thompson Sampling Router available from Docker Hub is [```seldonio/mab_Thompson_sampling:0.7```](https://hub.docker.com/r/seldonio/mab_Thompson_sampling).

## Wrap using s2i
### Persistence
For routers like multi-armed bandits it can be important to save the state after some learning has been done to avoid cold starts when re-deploying an inference graph. This can be achieved by setting ```PERSISTENCE=1``` in the ```.s2i/environment``` file before wrapping the source code. This will use redis to periodically save state of the component on the Seldon Core cluster.

### Wrap
```bash
make build
```

## Test the wrapped image

To test the generated docker image using the Seldon Core [internal API](../../../docs/reference/internal-api.md) refer to the [API testers](../../../docs/api-testing.md).
