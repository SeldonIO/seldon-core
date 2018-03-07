# Seldon-core Benchmarking

This page is a work in progress to provide benchmarking stats for  seldon-core. Please add further ideas and suggestions as an issue. 

## Goals

 * Load test REST and gRPC endpoints
 * Provide stability tests under load
 * Comparison to alternatives.

## Components

 * We use [locust](https://locust.io/) as our benchmarking tool.
 * We use Google Cloud Platform for the infrastructure to run kubernetes.


# Tests

## Maximum Throughput
To gauge the maximum throughput we will:

 * Call the seldon engine component directly thereby ignoring the additional latency that would be introduced by an external reverse proxy (Ambassador) or using the built in seldon API Front-End Oauth2 component.
 * Utilize a "stub" model that does nothing but return a hard-wired result from inside the engine.

This test will illustrate the maximum number of requests that can be pushed through seldon-core engine (which controls the request-response flow) as well as the added latency for the processing of REST and gRPC requests, e.g. serialization / deserialization.

We will use cordened off kubernetes nodes running locust so the latency from node to node prediction calls on GCP will also be part of the returned statistics.

A [notebook](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/benchmark_simple_model.ipynb) provides the end to end test for reproducability.

We use:

   * 1 replica of the stub model running on 1 n1-standard-16 GCP node
   * We use 3 nodes to run 64 locust slaves with a total of 256 clients calling as fast as they can.

See [notebook](https://github.com/SeldonIO/seldon-core/blob/master/notebooks/benchmark_simple_model.ipynb) for details.

### REST Results

A throughput of 12,000 request per second with average response time of 9ms is obtained.

|Method|Name|# requests|Requests/s|# failures|Median response time|Average response time|Min response time|Max response time|Average Content Size|
|--|--|--|--|--|--|--|--|--|--|
|POST|predictions|2363484|12088.95|0|4|9|1|5071|335|

With percentiles:

|Name|# requests|50%|66%|75%|80%|90%|95%|98%|99%|100%|
|--|--|--|--|--|--|--|--|--|--|--|
|POST predictions|2363484|4|5|7|9|28|43|60|69|5100|

### gRPC Results

A throughput of 28,000 requests per second with average response time of 1ms is obtained.

|Method|Name|# requests|Requests/s|# failures|Median response time|Average response time|Min response time|Max response time|Average Content Size|
|--|--|--|--|--|--|--|--|--|--|
|grpc|loadtest:5001|4622728|28256.39|0|1|1|0|5020|0|

With percentiles:

|Name|# requests|50%|66%|75%|80%|90%|95%|98%|99%|100%|
|--|--|--|--|--|--|--|--|--|--|--|
|grpc loadtest:5001|4622728|1|2|3|3|4|5|6|6|5000|
