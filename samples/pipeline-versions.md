## Seldon V2 Non Kubernetes Pipeline Version Updates



```python
!which seldon
```

    /home/clive/seldon/scv2/seldon-core-v2/operator/bin/seldon


### Model Join

Join two flows of data from two models as input to a third model.


```python
!seldon model load -f ./models/add10.yaml 
!seldon model load -f ./models/mul10.yaml 
```

    {}
    {}



```python
!seldon model status add10 -w ModelAvailable | jq -M .
!seldon model status mul10 -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!cat ./pipelines/version-test-a.yaml
```

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



```python
!seldon pipeline load -f ./pipelines/version-test-a.yaml
```

    {}



```python
!seldon pipeline status version-test -w PipelineReady | jq -M .
```

    {
      "pipelineName": "version-test",
      "versions": [
        {
          "pipeline": {
            "name": "version-test",
            "uid": "cdqjjkpqa12c739ab3rg",
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
            "lastChangeTimestamp": "2022-11-16T19:28:19.459332175Z"
          }
        }
      ]
    }



```python
!seldon pipeline infer version-test --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

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



```python
!cat ./pipelines/version-test-b.yaml
```

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



```python
!seldon pipeline load -f ./pipelines/version-test-b.yaml
```

    {}



```python
!seldon pipeline status version-test -w PipelineReady | jq -M .
```

    {
      "pipelineName": "version-test",
      "versions": [
        {
          "pipeline": {
            "name": "version-test",
            "uid": "cdqjjmpqa12c739ab3s0",
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
            "lastChangeTimestamp": "2022-11-16T19:28:27.768169880Z"
          }
        }
      ]
    }



```python
!seldon pipeline infer version-test --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

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



```python
!seldon pipeline load -f ./pipelines/version-test-a.yaml
```

    {}



```python
!seldon pipeline status version-test -w PipelineReady | jq -M .
```

    {
      "pipelineName": "version-test",
      "versions": [
        {
          "pipeline": {
            "name": "version-test",
            "uid": "cdqjjo9qa12c739ab3sg",
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
            "lastChangeTimestamp": "2022-11-16T19:28:33.139405433Z"
          }
        }
      ]
    }



```python
!seldon pipeline infer version-test --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

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



```python
!seldon pipeline unload version-test
```

    {}



```python
!seldon model unload add10
!seldon model unload mul10
```

    {}
    {}



```python

```
