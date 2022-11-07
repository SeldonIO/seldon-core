## Seldon V2 Non Kubernetes Pipeline Version Updates



```bash
which seldon
```
```json
    /home/clive/work/scv2/seldon-core-v2/operator/bin/seldon
```
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
            "uid": "cc5mrl9qre8m6toksbcg",
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
            "lastChangeTimestamp": "2022-08-28T13:29:58.003183799Z"
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
      ],
      "rawOutputContents": [
        "AAAwQQAAQEEAAFBBAABgQQ=="
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
            "uid": "cc5msjhqre8m6toksbd0",
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
            "lastChangeTimestamp": "2022-08-28T13:31:58.421910021Z"
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
      ],
      "rawOutputContents": [
        "AAAgQQAAoEEAAPBBAAAgQg=="
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
            "uid": "cc5mt1hqre8m6toksbdg",
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
            "lastChangeTimestamp": "2022-08-28T13:32:54.897566298Z"
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
      ],
      "rawOutputContents": [
        "AAAwQQAAQEEAAFBBAABgQQ=="
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
