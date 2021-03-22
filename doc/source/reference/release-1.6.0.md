# Seldon Core Release 1.6.0

A summary of the main contributions to the [Seldon Core release 1.6.0](https://github.com/SeldonIO/seldon-core/releases/tag/v1.6.0).

## MultiArmed Bandit Capabilities and Fixes

This release added further capabilities to router-enabled use-cases, such as the multi-armed bandits implementations in the seldon core repository. The extensions added in this release enable for all the `send_feedback` requests to be sent to the `router` component without explicit configuration required. With this change deploying the multi-armed bandits can be done without any further parameters in the CRD yaml.

An example of an Epsilon Greedy multi-armed bandit that routes the traffic to a sklearn and a xgboost model based in their performance can be found below:

```yaml
piVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: eg-experiment
spec:
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/credit_default_rf_model:0.2
          name: rf-model
        - image: seldonio/credit_default_xgb_model:0.2
          name: xgb-model
        - image: seldonio/mab_epsilon_greedy:1.6.0-dev
          name: eg-router
    graph:
      children:
      - name: rf-model
        type: MODEL
      - name: xgb-model
        type: MODEL
      name: eg-router
      parameters:
      - name: n_branches
        type: INT
        value: '2'
      - name: epsilon
        type: FLOAT
        value: '0.1'
      - name: verbose
        type: BOOL
        value: '1'
      - name: branch_names
        type: STRING
        value: rf:xgb
      - name: seed
        type: INT
        value: '1'
      type: ROUTER
    name: eg-2
    replicas: 1
    svcOrchSpec:
      env:
      - name: SELDON_ENABLE_ROUTING_INJECTION
        value: 'true'
```

You can find a full end to end example of the multi-armed bandits [implementation here](https://docs.seldon.io/projects/seldon-core/en/latest/analytics/routers.html).

## Added Github Actions for CI

We have now moved all our main unit tests in the CI to the Github Actions workflows. This makes the user experience much smoother for our community, ensuring that any contributors can see the logs of the tests in their PRs in real time. The integration tests using KIND are still run in our Jenkins X cluster due to the large memory requirements, but we will also be looking to transition these eventually.

## Updated Drift, Outlier and Explanations with Alibi

Our [CIFAR10 outlier detection example](../examples/drift_cifar10.html) and [CIFAR10 drift detection example](../examples/drift_cifar10.html) have been updated and now requires KNative 0.18 with a compatible istio (tested on 1.7.3). This utilizes the v1 resources of KNative Eventing to show off asnychronous outlier and drift detection on images.

Our explanations examples using [Alibi:Explain](https://github.com/SeldonIO/alibi) have been updated to the latest 0.4.3 release.

## Deprecation of Java Service Orchestrator (aka the Engine)

After the release of the Golang Service Orchestrator since version 1.0.0, the community has transition to using this more efficient component, and we have also been able to ensure we have full feature completeness (or relevant migration path) to ensure all the key functionality is present.

The Golang based service orchestrator brings a lot of benefits including higher throughputs and lower latencies. Being built in golang it is also able to leverage some of the core components that are developed in the Operator code.

This will also help remove any ambiguity around what component we refer to when we talk about the service orchestator - as often community members may get confused when hearing about "The Java Engine" or "The Golang Executor", now there is just one service orchestrator.

## Other highlights

* Seldon Operator now runs as non-root by default (with Security context override available)
* Resolved PyYAML CVE from Python base image
* Added support for V2 Protocol in outlier and drift detectors
* Handling V2 Protocol in request logger



