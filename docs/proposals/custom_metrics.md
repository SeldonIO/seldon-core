# Custom Metrics

## Summary

Allow users to easily add custom metrics to their Seldon Core components. For example, pass back extra metrics from a wrapped python model that can be collected by Prometheus and displayed om Grafana dashboards.

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


