# Performance tests
This section describes how a user can run performance tests to understand the limits of a particular SCv2 deployment.

The base directly is `tests/k6`

## Driver
[k6](https://k6.io/) is used to drive requests for load, unload and infer workloads. It is recommended that the load
test is run within the same cluster that has SCv2 installed as it requires internal access to some of the services
that are not automatically exposed to the outside world. Furthermore having the driver withthin the same cluster
minimises link latency to SCv2 entrypoint; therefore infer latencies are more representatives of actual overheads of the system.

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
  ```yaml
    # tests/k6/configs/k8s/base/k6.yaml
    args: [
      "--no-teardown",
      "--summary-export",
      "results/base.json",
      "--out",
      "csv=results/base.gz",
      "-u",
      "5",
      "-i",
      "100000",
      "-d",
      "120m",
      "scenarios/infer_constant_vu.js",
      ]
    # # infer_constant_rate
    # args: [
    #   "--no-teardown",
    #   "--summary-export",
    #   "results/base.json",
    #   "--out",
    #   "csv=results/base.gz",
    #   "scenarios/infer_constant_rate.js",
    #   ]
    # # k8s-test-script
    # args: [
    #   "--summary-export",
    #   "results/base.json",
    #   "--out",
    #   "csv=results/base.gz",
    #   "scenarios/k8s-test-script.js",
    #   ]
    # # core2_qa_control_plane_ops
    # args: [
    #   "--no-teardown",
    #   "--verbose",
    #   "--summary-export",
    #   "results/base.json",
    #   "--out",
    #   "csv=results/base.gz",
    #   "-u",
    #   "5",
    #   "-i",
    #   "10000",
    #   "-d",
    #   "9h",
    #   "scenarios/core2_qa_control_plane_ops.js",
    #   ]
   ```
for a full list, check [k6 args](https://k6.io/docs/using-k6/options/)
* Environment variables
  ```yaml
    - name: INFER_HTTP_ITERATIONS
      value: "1"
    - name: INFER_GRPC_ITERATIONS
      value: "1"
    - name: MODELNAME_PREFIX
      value: "tfsimplea,pytorch-cifar10a,tfmnista,mlflow-winea,irisa"
    - name: MODEL_TYPE
      value: "tfsimple,pytorch_cifar10,tfmnist,mlflow_wine,iris"
    # Specify MODEL_MEMORY_BYTES using unit-of measure suffixes (k, M, G, T)
    # rather than numbers without units of measure. If supplying "naked
    # numbers", the seldon operator will take care of converting the number
    # for you but also take ownership of the field (as FieldManager), so the
    # next time you run the scenario creating/updating of the model CR will
    # fail.
    - name: MODEL_MEMORY_BYTES
      value: "400k,8M,43M,200k,3M"
    - name: MAX_MEM_UPDATE_FRACTION
      value: "0.1"
    - name: MAX_NUM_MODELS
      value: "800,100,25,100,100"
      # value: "0,0,25,100,100"
    #
    # MAX_NUM_MODELS_HEADROOM is a variable used by control-plane tests.
    # It's the approximate number of models that can be created over
    # MAX_NUM_MODELS over the experiment. In the worst case scenario
    # (very unlikely) the HEADROOM values may temporarily exceed the ones
    # specified here with the number of VUs, because each VU checks the
    # headroom constraint independently before deciding on the available
    # operations (no communication/sync between VUs)
    # - name: MAX_NUM_MODELS_HEADROOM
    #   value: "20,5,0,20,30"
    #
    # MAX_MODEL_REPLICAS is used by control-plane tests. It controls the
    # maximum number of replicas that may be requested when
    # creating/updating models of a given type.
    # - name: MAX_MODEL_REPLICAS
    #   value: "2,2,0,2,2"
    #
    - name: INFER_BATCH_SIZE
      value: "1,1,1,1,1"
    # MODEL_CREATE_UPDATE_DELETE_BIAS defines the probability ratios between
    # the operations, for control-plane tests. For example, "1, 4, 3"
    # makes an Update four times more likely then a Create, and a Delete 3
    # times more likely than the Create.
    # - name: MODEL_CREATE_UPDATE_DELETE_BIAS
    #   value: "1,3,1"
    - name: WARMUP
      value: "false"
  ```
    * for `MODEL_TYPE`, choose from:
  ```js
  // tests/k6/components/model.js
    import { dump as yamlDump } from "https://cdn.jsdelivr.net/npm/js-yaml@4.1.0/dist/js-yaml.mjs";
    import { getConfig } from '../components/settings.js'

    const tfsimple_string = "tfsimple_string"
    const tfsimple = "tfsimple"
    const iris = "iris"  // mlserver
    const pytorch_cifar10 = "pytorch_cifar10"
    const tfmnist = "tfmnist"
    const tfresnet152 = "tfresnet152"
    const onnx_gpt2 = "onnx_gpt2"
    const mlflow_wine = "mlflow_wine" // mlserver
    const add10 = "add10" // https://github.com/SeldonIO/triton-python-examples/tree/master/add10
    const sentiment = "sentiment" // mlserver
  ```
