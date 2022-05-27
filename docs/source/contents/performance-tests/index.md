# Performance tests
This section describes how a user can run performance tests to understand the limits of a particular SCv2 deployement.

The base directly is `tests/k6`

## Driver
[k6](https://k6.io/) is used to drive requests for load, unload and infer workloads. It is recommended that the load test is run withing the same cluster that has SCv2 installed as it requires internal access to some of the services that are not automatically exposed to the outside world. Furthermore having the driver withthin the same cluster minimises link latency to SCv2 entrypoint; therefore infer latencies are more representives of actual overheads of the system.

## Tests

* Envoy
Tests synchronous inference requests via envoy

To run: `make deploy-envoy-test`

* Agent
Tests inference requests direct to a specific agent, defaults to triton-0 or mlserver-0

To run: `make deploy-rproxy-test` pr `make deploy-rproxy-mlserver-test`

* Server
Tests inference requests direct to a specific server (bypassing agent), defaults to triton-0 or mlserver-0

to run: `make deploy-server-test` or `deploy-server-mlserver-test`

* Pipeline gateway (HTTP-Kafka gateway)
Tests inference requests to one-node pipeline HTTP and GPRC requests

To run: `make deploy-kpipeline-test`

* Model gateway (Kafka-HTTP gateway)
Tests inference requests to a model via kafka

To run: `deploy-kmodel-test`

## Results

One way to look at results is to look at the log of the pod that executed the kubernetes job.

Results can also be persisted to a gs bucket, a service account `k6-sa-key` in the same namespace is required,

Users can also look at the metrics that are exposed in prometheus while the test is underway

## Building k6 image

In the case a user is modifying the actual scenario of the test:

* `export DOCKERHUB_USERNAME=mydockerhubaccount`
* build the k6 image via `make build-push`
* in the same shell environment, deploying jobs will use this custome built docker image

## Modifying tests

Users can modify settings of the tests in `tests/k6/configs/k8s/base/k6.yaml`. This will apply to all subsequent tests that are deployed using the above process.

## Settings

Some settings that can be changed

* k6 args
  ```{literalinclude} ../../../../tests/k6/configs/k8s/base/k6.yaml
   :language: yaml
   :start-after: args
   :end-before: env
   ```
for a full list, check [k6 args](https://k6.io/docs/using-k6/options/)
* Environment variables
  ```{literalinclude} ../../../../tests/k6/configs/k8s/base/k6.yaml
   :language: yaml
   :start-after: SCHEDULER_ENDPOINT}:9004
   :end-before: GOOGLE_APPLICATION_CREDENTIALS 
   ```
    * for `MODEL_TYPE`, choose from: 
  ```{literalinclude} ../../../../tests/k6/components/model.js
   :language: javascript
   :end-before: const models
   ```
