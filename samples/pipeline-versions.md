## Seldon V2 Non Kubernetes Pipeline Version Updates

### Model Join

Join two flows of data from two models as input to a third model.

```bash
seldon model load -f ./models/add10.yaml
seldon model load -f ./models/mul10.yaml
```

```json
{}
{}

```

```bash
seldon model status add10 -w ModelAvailable | jq -M .
seldon model status mul10 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

```bash
cat ./pipelines/version-test-a.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: version-test
spec:
  steps:
    - name: add10
  output:
    steps:
    - add10

```

```bash
seldon pipeline load -f ./pipelines/version-test-a.yaml
```

```json
{}

```

```bash
seldon pipeline status version-test -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "version-test",
  "versions": [
    {
      "pipeline": {
        "name": "version-test",
        "uid": "cg5g7ck6dpcs73c4qho0",
        "version": 1,
        "steps": [
          {
            "name": "add10"
          }
        ],
        "output": {
          "steps": [
            "add10.outputs"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-10T10:17:22.883632333Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer version-test --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```

```bash
cat ./pipelines/version-test-b.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: version-test
spec:
  steps:
    - name: mul10
  output:
    steps:
    - mul10

```

```bash
seldon pipeline load -f ./pipelines/version-test-b.yaml
```

```json
{}

```

```bash
seldon pipeline status version-test -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "version-test",
  "versions": [
    {
      "pipeline": {
        "name": "version-test",
        "uid": "cg5g7ek6dpcs73c4qhog",
        "version": 2,
        "steps": [
          {
            "name": "mul10"
          }
        ],
        "output": {
          "steps": [
            "mul10.outputs"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 2,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-10T10:17:30.576466739Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer version-test --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          10,
          20,
          30,
          40
        ]
      }
    }
  ]
}

```

```bash
seldon pipeline load -f ./pipelines/version-test-a.yaml
```

```json
{}

```

```bash
seldon pipeline status version-test -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "version-test",
  "versions": [
    {
      "pipeline": {
        "name": "version-test",
        "uid": "cg5g7g46dpcs73c4qhp0",
        "version": 3,
        "steps": [
          {
            "name": "add10"
          }
        ],
        "output": {
          "steps": [
            "add10.outputs"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 3,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-10T10:17:36.568170759Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer version-test --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```

```bash
seldon pipeline unload version-test
```

```json
{}

```

```bash
seldon model unload add10
seldon model unload mul10
```

```json
{}
{}

```

```python

```
