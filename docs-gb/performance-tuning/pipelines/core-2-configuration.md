---
description: Reducing Data Processing Overheads in Seldon Core 2.
---

## Reducing Data Processing Overheads

When tuning performance for pipelines, reducing the overhead of Core 2 components responsible for data-processing within a pipeline is another aspect to consider. In Core 2, four components influence this overhead:

- `pipelinegateway`  which handles pipeline requests
- `modelgateway`  which sends requests to model inference servers
- `dataflow-engine`  runs Kafka KStream topologies to manage data streamed between models in a pipeline
- the Kafka cluster

Given that Core 2 uses Kafka as the messaging system to communicate data between models, lowering the network latency between Core 2 and Kafka (especially when the Kafka installation is in a separate cluster to Core) will improve pipeline performance. 

Additionally, the number of Kafka partitions per topic (which must be fixed across all models in a pipeline) significantly influences

1. Kafka’s maximum throughput, and
2. the effective number of replicas for `pipelinegateway`, `dataflow-engine` and `modelgateway` 

As a baseline for serving a high inferencing RPS across multiple pipelines, we recommend using as many replicas of `pipelinegateway` and `dataflow-engine` as you have Kafka topic partitions in order to leverage the balanced distribution of inference traffic. In this case, each `dataflow-engine` will process the data from one partition, across all pipelines. Increasing the number of dataflow-engine replicas further starts sharding the pipelines across the available replicas, with each pipeline being processed by `maxShardCountMultiplier` replicas (see detailed pipeline scalability docs for configuration details)

Similarly, `modelgateway` can handle more throughput if its number of workers and number of replicas is increased. `modelgateway` has two scalability parameters that can be set via environment variables:

- `MODELGATEWAY_NUM_WORKERS`
- `MODELGATEWAY_MAX_NUM_CONSUMERS`

Each model within a Kubernetes namespace is consistently assigned to one modelgateway consumer (based on their index in a hash table of size `MODELGATEWAY_MAX_NUM_CONSUMERS`); The size of the hash table influences how many models will share the same consumer.

For each consumer, a `MODELGATEWAY_NUM_WORKERS` number of lightweight inference workers (goroutines) are created to forward requests to the inference servers and wait for responses.

Increasing these parameters (starting with an increase in the number of workers) will improve throughput if the `modelgateway` pod has enough resources to support more workers and consumers.
