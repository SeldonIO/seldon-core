# Operational Metrics

While the system is running we collect metrics via Prometheus that allow users to observe different
aspects of SCv2 with regards to throughput, latency, memory, CPU etc. This is in addition to the standard
Kubernetes metrics that are scraped by Prometheus. There is a also a Grafana dashboard (referenced below)
that provides an overview of the system.

## List of SCv2 metrics

The list of SCv2 metrics that we are compiling is as follows.

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

Many of these metrics are model and pipeline level counters and gauges. We also aggregate some of
these metrics to speed up the display of graphs. We don't presently store per-model histogram metrics
for performance reasons. However, we do presently store per-pipeline histogram metrics.

This is experimental and these metrics are bound to change to reflect the trends we want to capture as
we get more information about the usage of the system.

## Grafana dashboard

We have a prebuilt Grafana dashboard that makes use of many of the metrics that we expose.

![kafka](../images/dashboard.png)

### Local Use

Grafana and Prometheus are available when you run Seldon locally. You will be able to connect to the Grafana
dashboard at `http://localhost:3000`. Prometheus will be available at `http://localhost:9090`.

### Kubernetes Installation

Download the dashboard from [SCv2 dashboard](https://github.com/SeldonIO/seldon-core/blob/v2/prometheus/dashboards/seldon.json)
and import it in Grafana, making sure that the data source is pointing to the correct Prometheus store.
Find more information on how to import the dashboard [here](https://grafana.com/docs/grafana/latest/dashboards/export-import/).

### Local Metrics Examples

An [example](./local-metrics-test.md) to show raw metrics that Prometheus will scrape.
