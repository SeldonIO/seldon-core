---
description: This page provides guidance about scaling Seldon Core 2 services
---

Seldon Core 2 runs with several control and dataplane components. The scaling of these resources is discussed below:

- **Pipeline gateway**: The pipeline gateway handles REST and gRPC synchronous requests to Pipelines. It is stateless and can be scaled based on traffic demand.
- **Model gateway**: This component pulls model requests from Kafka and sends them to inference servers. It can be scaled up to the partition factor of your Kafka topics. At present we set a uniform partition factor for all topics in one installation of Seldon.
- **Dataflow engine**: The dataflow engine runs KStream topologies to manage Pipelines. It can run as multiple replicas and the scheduler will balance Pipelines to run across it with a consistent hashing load balancer. Each Pipeline is managed up to the partition factor of Kafka (presently hardwired to one). We recommend using as many replicas of dataflow-engine as you have Kafka partitions in order to leverage the balanced distribution of inference traffic using hashing
- **Scheduler**: The scheduler manages the control plane operations. It is presently required to be one replica as it maintains internal state within a BadgerDB held on local persistent storage (stateful set in Kubernetes). Performance tests have shown this not to be a bottleneck at present.
- **Kubernetes Controller**: The Kubernetes controller manages resources updates on the cluster which it passes on to the Scheduler. It is by default one replica but has the ability to scale.
- **Envoy**: Envoy replicas get their state from the scheduler for routing information and can be scaled as needed.

