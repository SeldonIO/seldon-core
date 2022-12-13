## Metric Examples

This notebook tests the exposed Prometheus metrics of model and pipeline servers.

Requires: `prometheus_client` and `requests` libraries.
See docs for full set of metrics available.

```python
mlserver_metrics_host="0.0.0.0:9006"
triton_metrics_host="0.0.0.0:9007"
pipeline_metrics_host="0.0.0.0:9009"
```

```python
from prometheus_client.parser import text_string_to_metric_families
import requests

def scrape_metrics(host):
    data = requests.get(f"http://{host}/metrics").text
    return {
        family.name: family for family in text_string_to_metric_families(data)
    }

def print_sample(family, label, value):
    for sample in family.samples:
        if sample.labels[label] == value:
            print(sample)

def get_model_infer_count(host, model_name):
    metrics = scrape_metrics(host)
    family = metrics["seldon_model_infer"]
    print_sample(family, "model", model_name)

def get_pipeline_infer_count(host, pipeline_name):
    metrics = scrape_metrics(host)
    family = metrics["seldon_pipeline_infer"]
    print_sample(family, "pipeline", pipeline_name)
```

### MLServer Model

```bash
seldon model load -f ./models/sklearn-iris-gs.yaml
seldon model status iris -w ModelAvailable | jq -M .
```

```json
{}
{}

```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```yaml
Success: map[:iris_1::50]

```

```bash
seldon model infer iris --inference-mode grpc -i 100 \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
```

```yaml
Success: map[:iris_1::100]

```

```python
get_model_infer_count(mlserver_metrics_host,"iris")
```

```yaml
Sample(name='seldon_model_infer_total', labels={'code': '200', 'method_type': 'rest', 'model': 'iris', 'model_internal': 'iris_1', 'server': 'mlserver', 'server_replica': '0'}, value=50.0, timestamp=None, exemplar=None)
Sample(name='seldon_model_infer_total', labels={'code': 'OK', 'method_type': 'grpc', 'model': 'iris', 'model_internal': 'iris_1', 'server': 'mlserver', 'server_replica': '0'}, value=100.0, timestamp=None, exemplar=None)

```

```bash
seldon model unload iris
```

```json
{}

```

### Triton Model

Load the model.

```bash
seldon model load -f ./models/tfsimple1.yaml
seldon model status tfsimple1 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

```bash
seldon model infer tfsimple1 -i 50\
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```

```yaml
Success: map[:tfsimple1_1::50]

```

```bash
seldon model infer tfsimple1 --inference-mode grpc -i 100 \
    '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
```

```yaml
Success: map[:tfsimple1_1::100]

```

```python
get_model_infer_count(triton_metrics_host,"tfsimple1")
```

```yaml
Sample(name='seldon_model_infer_total', labels={'code': '200', 'method_type': 'rest', 'model': 'tfsimple1', 'model_internal': 'tfsimple1_1', 'server': 'triton', 'server_replica': '0'}, value=50.0, timestamp=None, exemplar=None)
Sample(name='seldon_model_infer_total', labels={'code': 'OK', 'method_type': 'grpc', 'model': 'tfsimple1', 'model_internal': 'tfsimple1_1', 'server': 'triton', 'server_replica': '0'}, value=100.0, timestamp=None, exemplar=None)

```

```bash
seldon model unload tfsimple1
```

```json
{}

```

### Pipeline

```bash
seldon model load -f ./models/tfsimple1.yaml
seldon model load -f ./models/tfsimple2.yaml
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon pipeline load -f ./pipelines/tfsimples.yaml
seldon pipeline status tfsimples -w PipelineReady
```

```json
{}
{}
{}
{}
{}
{"pipelineName":"tfsimples", "versions":[{"pipeline":{"name":"tfsimples", "uid":"cdqji39qa12c739ab3o0", "version":2, "steps":[{"name":"tfsimple1"}, {"name":"tfsimple2", "inputs":["tfsimple1.outputs"], "tensorMap":{"tfsimple1.outputs.OUTPUT0":"INPUT0", "tfsimple1.outputs.OUTPUT1":"INPUT1"}}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":2, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-11-16T19:25:01.255955114Z"}}]}

```

```bash
seldon pipeline infer tfsimples -i 50 \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```

```yaml
Success: map[:tfsimple1_1::50 :tfsimple2_1::50 :tfsimples.pipeline::50]

```

```python
get_pipeline_infer_count(pipeline_metrics_host,"tfsimples")
```

```yaml
Sample(name='seldon_pipeline_infer_total', labels={'code': '200', 'method_type': 'rest', 'pipeline': 'tfsimples', 'server': 'pipeline-gateway'}, value=50.0, timestamp=None, exemplar=None)

```

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
seldon pipeline unload tfsimples
```

```json
{}
{}
{}

```

```python

```
