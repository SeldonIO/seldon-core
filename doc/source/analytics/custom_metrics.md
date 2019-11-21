# Custom Metrics

Seldon Core exposes basic metrics via Prometheus endpoints on its service orchestrator that include request count, request time percentiles and rolling accuracy for each running model. However, you may wish to expose custom metrics from your components which are automatically added to Prometheus. For this purpose you can supply extra fields in the returned meta data of the response object in the API calls to your components as illustrated below:

```json
{
	"meta": {
		"metrics": [
			{
				"type": "COUNTER",
				"key": "mycounter",
				"value": 1.0,
				"tags": {"mytag":"mytagvalue"}
			},
			{
				"type": "GAUGE",
				"key": "mygauge",
				"value": 22.0
			},
			{
				"type": "TIMER",
				"key": "mytimer",
				"value": 1.0
			}
		]
	},
	"data": {
		"ndarray": [
			[
				1,
				2
			]
		]
	}
}
```

We provide three types of metric that can be returned in the meta.metrics list:

 * COUNTER : a monotonically increasing value. It will be added to any existing value from the metric key.
 * GAUGE : an absolute value showing a level, it will overwrite any existing value.
 * TIMER : a time value (in msecs).

Each metric, apart from the type, takes a key and a value. The proto buffer definition is shown below:

```protobuf
message Metric {
 enum MetricType {
     COUNTER = 0;
     GAUGE = 1;
     TIMER = 2;
 }
 string key = 1;
 MetricType type = 2;
 float value = 3;
 map<string,string> tags = 4; 
}
```


As we expose the metrics via Prometheus, if ```tags``` are added they must appear in every metric response and always have the same set of keys since Prometheus does not allow metrics to have varying numbers of tags. This condition is enforced by the [micrometer](https://micrometer.io/) library we use to expose the metrics. Exceptions will happen if this condition is violated.

At present the following Seldon Core wrappers provide integrations with custom metrics:

 * [Python Wrapper](../python/index.html)


## Example

There is an [example notebook illustrating a model with custom metrics in python](../examples/tmpl_model_with_metrics.html).
