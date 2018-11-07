# Custom Metrics

## Summary

Allow users to easily add custom metrics to their Seldon Core components. For example, pass back extra metrics from a wrapped python model that can be collected by Prometheus and displayed on Grafana dashboards.

## Proposal

Extend the SeldonMessage proto buffer to have a "metrics" element in the meta data part, e.g.,

```JSON
{
"meta" : {
  "metrics" : [
    { "key" : "my_metric_1", "type" : "counter", "value" : 1 },
    { "key" : "my_metric_2", "type" : "guage", "value" : 223 }
  ]
}
}
```

These metrics would be automaticaly exposed to prometheus from the Seldon Orchestrator Engine.

The wrappers would need to be updated to allow users to not just return a prediction but also optionally provide metrics to return.

## Metrics definition

The extended meta data section would be:

```
message Meta {
  string puid = 1; 
  map<string,google.protobuf.Value> tags = 2;
  map<string,int32> routing = 3;
  map<string,string> requestPath = 4;
  repeated Metric metrics = 5;
}

message Metric {
 enum MetricType {
   COUNTER = 0;
   GAUGE = 1;
   TIMER = 2;
 }
 string key = 1;
 MetricType type = 2;
 float value = 3;
 string graphId = 4;
}
```


## Metric Types - Histogram Complexities
We use [Micrometer](https://micrometer.io) for exposing metrics. Counter and gauge are pretty standard but Prometheus has Histogram and Summary. Histogram seems most advantagous as you can summarize the data on prometheus later. However, you need to set the number of buckets you want to collect statistics. For micrometer the default is to set a min and max range and it will create a set of buckets for you. For Micrometer Timers there are defaults in [Micrometer](https://micrometer.io/docs/concepts#_histograms_and_percentiles) set for the range 1ms - 1minute. The trouble with general histograms is you would need to expose this setting which is probably too complex.

 * Suggest we support just "TIMER" which is essentially a Prometheus histogram for durations with a  range 1ms-1minute and a default set of buckets


## Engine Implementation

 1. For each component if there is a metrics section parse and expose via prometheus each metric of the appropriate type.
 2. Merge all metrics into final set for returning externally adding graph id of the component that returned the metrics if missing.

## Wrapper Implementations

### Python Wrapper

Add optional new function in class user defines

```
def metrics(self):
  return [
    { "key" : "my_metric_1", "type" : "counter", "value" : self.counter1 },
    { "key" : "my_metric_2", "type" : "guage", "value" : self.guage1 }
  ]
```


