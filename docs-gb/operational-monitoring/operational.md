---
description: Learn how to monitor and analyze operational metrics in Seldon Core 2 using Prometheus and Grafana. This comprehensive guide covers model and pipeline performance metrics, resource utilization tracking, cache management, memory monitoring, and best practices for collecting and visualizing ML serving metrics in production environments.
---

# Operational Metrics

While the system runs, Prometheus collects metrics that enable you to observe various aspects of Seldon Core 2, including throughput, latency, memory, and CPU usage. In addition to the standard Kubernetes metrics scraped by Prometheus, a [Grafana dashboard](observability.md/#grafana) provides a comprehensive system overview.

## List of Seldon Core 2 metrics

The list of Seldon Core 2 metrics that are compiling is as follows.

For the agent that sits next to the inference servers:

```go
// scheduler/pkg/metrics/agent.go
const (
	// Histograms do no include pipeline label for efficiency
	modelHistogramName = "seldon_model_infer_api_seconds"
	// We use base infer counters to store core metrics per pipeline
	modelInferCounterName                 = "seldon_model_infer_total"
	modelInferLatencyCounterName          = "seldon_model_infer_seconds_total"
	modelAggregateInferCounterName        = "seldon_model_aggregate_infer_total"
	modelAggregateInferLatencyCounterName = "seldon_model_aggregate_infer_seconds_total"
)

// Agent metrics
const (
	cacheEvictCounterName                              = "seldon_cache_evict_count"
	cacheMissCounterName                               = "seldon_cache_miss_count"
	loadModelCounterName                               = "seldon_load_model_counter"
	unloadModelCounterName                             = "seldon_unload_model_counter"
	loadedModelGaugeName                               = "seldon_loaded_model_gauge"
	loadedModelMemoryGaugeName                         = "seldon_loaded_model_memory_bytes_gauge"
	evictedModelMemoryGaugeName                        = "seldon_evicted_model_memory_bytes_gauge"
	serverReplicaMemoryCapacityGaugeName               = "seldon_server_replica_memory_capacity_bytes_gauge"
	serverReplicaMemoryCapacityWithOverCommitGaugeName = "seldon_server_replica_memory_capacity_overcommit_bytes_gauge"
)
```

For the pipeline gateway that handles requests to pipelines:

```go
// scheduler/pkg/metrics/gateway.go
// The aggregate metrics exist for efficiency, as the summation can be
// very slow in Prometheus when many pipelines exist.
const (
	// Histograms do no include model label for efficiency
	pipelineHistogramName = "seldon_pipeline_infer_api_seconds"
	// We use base infer counters to store core metrics per pipeline
	pipelineInferCounterName                 = "seldon_pipeline_infer_total"
	pipelineInferLatencyCounterName          = "seldon_pipeline_infer_seconds_total"
	pipelineAggregateInferCounterName        = "seldon_pipeline_aggregate_infer_total"
	pipelineAggregateInferLatencyCounterName = "seldon_pipeline_aggregate_infer_seconds_total"
)
```

Many of these metrics are model and pipeline level counters and gauges. Some of these metrics are aggregated to speed up the display of graphs. Currently,per-model histogram metrics are not stored for performance reasons. However, per-pipeline histogram metrics are stored.

This is experimental, and these metrics are expected to evolve to better capture relevant trends as more information becomes available about system usage.


### Local Metrics Examples

An [example](./local-metrics-test.md) to show raw metrics that Prometheus will scrape.
