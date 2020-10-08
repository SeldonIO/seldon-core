# Seldon-core Benchmarking and Load Testing

This page is a work in progress to provide benchmarking and load testing.

This work is ongoing and we welcome feedback

## Tools

 * For REST tests we use [vegeta](https://github.com/tsenart/vegeta)
 * For gRPC tests we use [ghz](https://ghz.sh/)

## Service Orchestrator

These benchmark tests are to evaluate the extra latency added by including the service orchestrator.

 * [Service orchestrator benchmark](../examples/bench_svcOrch.html)

### Results

On A 3 node DigitalOcean cluster 24vCPUs 96 GB, running Tensorflow Flowers image classfier.

| Test | Additional latency |
| ---  | ------------------ |
| REST | 9ms |
| gRPC | 4ms |

Further work:

 * Statistical confidence test


## Tensorflow

Test the max throughput and HPA usage.

 * [Tensorflow benchmark](../examples/bench_tensorflow.html)

### Results

On A 3 node DigitalOcean cluster 24vCPUs 96 GB, running Tensorflow Flowers image classfier with HPA and running at max throughput for a single model. No ramp up, as vegeta does not support this. See notebook for details.

```
Latencies:

mean: 259.990239 ms
50th: 131.917169 ms
90th: 310.053255 ms
95th: 916.684759 ms
99th: 2775.05271 ms

Throughput: 23.997572337989126/s
Errors: False
```

## Flexible Benchmarking with Argo Workflows

We have also an example that shows how to leverage the batch processing workflow that we showcase in the examples, but to perform benchmarking with Seldon Core models.

 * [Seldon deployment benchmark](../examples/vegeta_bench_argo_workflows.html)

