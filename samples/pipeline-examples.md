## Seldon V2 Non Kubernetes Pipeline Examples



```bash
which seldon
```

    /home/clive/work/scv2/seldon-core-v2/operator/bin/seldon
```
````
### Model Chaining

Chain the output of one model into the next. Also shows chaning the tensor names via `tensorMap` to conform to the inputs of the second model.


```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
```
````

```bash
seldon model load -f ./models/tfsimple1.yaml 
seldon model load -f ./models/tfsimple2.yaml 
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
cat ./pipelines/tfsimples.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimples
      namespace: seldon-mesh
    spec:
      steps:
        - name: tfsimple1
        - name: tfsimple2
          inputs:
          - tfsimple1
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT0
            tfsimple1.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - tfsimple2
```
````

```bash
seldon pipeline load -f ./pipelines/tfsimples.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status tfsimples -w PipelineReady| jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "tfsimples",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimples",
            "uid": "cahec4fu7k0kl6oagsp0",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple1"
              },
              {
                "name": "tfsimple2",
                "inputs": [
                  "tfsimple1.outputs"
                ],
                "tensorMap": {
                  "tfsimple1.outputs.OUTPUT0": "INPUT0",
                  "tfsimple1.outputs.OUTPUT1": "INPUT1"
                }
              }
            ],
            "output": {
              "steps": [
                "tfsimple2.outputs"
              ]
            },
            "kubernetesMeta": {
              "namespace": "seldon-mesh"
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-06-10T06:34:58.712206570Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer tfsimples \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "model_name": "",
      "outputs": [
        {
          "data": null,
          "name": "OUTPUT0",
          "shape": [
            1,
            16
          ],
          "datatype": "INT32"
        },
        {
          "data": null,
          "name": "OUTPUT1",
          "shape": [
            1,
            16
          ],
          "datatype": "INT32"
        }
      ],
      "rawOutputContents": [
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="
      ]
    }
```
````

```bash
seldon pipeline infer tfsimples --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "outputs": [
        {
          "name": "OUTPUT0",
          "datatype": "INT32",
          "shape": [
            "1",
            "16"
          ],
          "contents": {
            "intContents": [
              2,
              4,
              6,
              8,
              10,
              12,
              14,
              16,
              18,
              20,
              22,
              24,
              26,
              28,
              30,
              32
            ]
          }
        },
        {
          "name": "OUTPUT1",
          "datatype": "INT32",
          "shape": [
            "1",
            "16"
          ],
          "contents": {
            "intContents": [
              2,
              4,
              6,
              8,
              10,
              12,
              14,
              16,
              18,
              20,
              22,
              24,
              26,
              28,
              30,
              32
            ]
          }
        }
      ],
      "rawOutputContents": [
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="
      ]
    }
```
````

```bash
seldon pipeline inspect tfsimples
```
````{collapse} Expand to see output
```json

    ---
    seldon.default.model.tfsimple1.inputs
    {"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    ---
    seldon.default.model.tfsimple1.outputs
    {"modelName":"tfsimple1_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawOutputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    ---
    seldon.default.model.tfsimple2.inputs
    {"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    ---
    seldon.default.model.tfsimple2.outputs
    {"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawOutputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    ---
    seldon.default.pipeline.tfsimples.inputs
    {"modelName":"tfsimples","inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    ---
    seldon.default.pipeline.tfsimples.outputs
    {"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawOutputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
```
````

```bash
seldon pipeline unload tfsimples
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
### Model Join

Join two flows of data from two models as input to a third model.


```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
cat ./models/tfsimple3.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple3
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
```
````

```bash
seldon model load -f ./models/tfsimple1.yaml 
seldon model load -f ./models/tfsimple2.yaml 
seldon model load -f ./models/tfsimple3.yaml 
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon model status tfsimple3 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````

```bash
cat ./pipelines/tfsimples-join.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: join
      namespace: seldon-mesh
    spec:
      steps:
        - name: tfsimple1
        - name: tfsimple2
        - name: tfsimple3      
          inputs:
          - tfsimple1.outputs.OUTPUT0
          - tfsimple2.outputs.OUTPUT1
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT0
            tfsimple2.outputs.OUTPUT1: INPUT1
      output:
        steps:
        - tfsimple3
```
````

```bash
seldon pipeline load -f ./pipelines/tfsimples-join.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status join -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "join",
      "versions": [
        {
          "pipeline": {
            "name": "join",
            "uid": "cahed2fu7k0kl6oagspg",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple1"
              },
              {
                "name": "tfsimple2"
              },
              {
                "name": "tfsimple3",
                "inputs": [
                  "tfsimple1.outputs.OUTPUT0",
                  "tfsimple2.outputs.OUTPUT1"
                ],
                "tensorMap": {
                  "tfsimple1.outputs.OUTPUT0": "INPUT0",
                  "tfsimple2.outputs.OUTPUT1": "INPUT1"
                }
              }
            ],
            "output": {
              "steps": [
                "tfsimple3.outputs"
              ]
            },
            "kubernetesMeta": {
              "namespace": "seldon-mesh"
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-06-10T06:36:58.313959175Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer join --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "outputs": [
        {
          "name": "OUTPUT0",
          "datatype": "INT32",
          "shape": [
            "1",
            "16"
          ],
          "contents": {
            "intContents": [
              2,
              4,
              6,
              8,
              10,
              12,
              14,
              16,
              18,
              20,
              22,
              24,
              26,
              28,
              30,
              32
            ]
          }
        },
        {
          "name": "OUTPUT1",
          "datatype": "INT32",
          "shape": [
            "1",
            "16"
          ],
          "contents": {
            "intContents": [
              2,
              4,
              6,
              8,
              10,
              12,
              14,
              16,
              18,
              20,
              22,
              24,
              26,
              28,
              30,
              32
            ]
          }
        }
      ],
      "rawOutputContents": [
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="
      ]
    }
```
````

```bash
seldon pipeline unload join
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
seldon model unload tfsimple3
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````
### Conditional

Shows conditional data flows - one of two models is run based on output tensors from first.


```bash
cat ./models/conditional.yaml
cat ./models/add10.yaml
cat ./models/mul10.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: conditional
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/conditional"
      requirements:
      - triton
      - python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton
      - python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/mul10"
      requirements:
      - triton
      - python
```
````

```bash
seldon model load -f ./models/conditional.yaml 
seldon model load -f ./models/add10.yaml 
seldon model load -f ./models/mul10.yaml 
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````

```bash
seldon model status conditional -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
seldon model status mul10 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````

```bash
cat ./pipelines/conditional.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-conditional
      namespace: seldon-mesh
    spec:
      steps:
      - name: conditional
      - name: mul10
        inputs:
        - conditional.outputs.OUTPUT0
        tensorMap:
          conditional.outputs.OUTPUT0: INPUT
      - name: add10
        inputs:
        - conditional.outputs.OUTPUT1
        tensorMap:
          conditional.outputs.OUTPUT1: INPUT
      output:
        steps:
        - mul10
        - add10
        stepsJoin: any
```
````

```bash
seldon pipeline load -f ./pipelines/conditional.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status tfsimple-conditional -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "tfsimple-conditional",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple-conditional",
            "uid": "cahedm7u7k0kl6oagsq0",
            "version": 1,
            "steps": [
              {
                "name": "add10",
                "inputs": [
                  "conditional.outputs.OUTPUT1"
                ],
                "tensorMap": {
                  "conditional.outputs.OUTPUT1": "INPUT"
                }
              },
              {
                "name": "conditional"
              },
              {
                "name": "mul10",
                "inputs": [
                  "conditional.outputs.OUTPUT0"
                ],
                "tensorMap": {
                  "conditional.outputs.OUTPUT0": "INPUT"
                }
              }
            ],
            "output": {
              "steps": [
                "mul10.outputs",
                "add10.outputs"
              ],
              "stepsJoin": "ANY"
            },
            "kubernetesMeta": {
              "namespace": "seldon-mesh"
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-06-10T06:38:17.710751776Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer tfsimple-conditional --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[0]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```
````{collapse} Expand to see output
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
````

```bash
seldon pipeline infer tfsimple-conditional --inference-mode grpc \
 '{"model_name":"outlier","inputs":[{"name":"CHOICE","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```
````{collapse} Expand to see output
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
````

```bash
seldon pipeline unload tfsimple-conditional
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload conditional
seldon model unload add10
seldon model unload mul10
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````
### Error
An example which errors when its arguments sum to greater than 100. Shows stopping a pipeline early on error.


```bash
cat ./models/outlier-error.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: outlier-error
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/outlier"
      requirements:
      - triton
      - python
```
````

```bash
seldon model load -f ./models/outlier-error.yaml 
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model status outlier-error -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
cat ./pipelines/error.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: error
    spec:
      steps:
        - name: outlier-error
      output:
        steps:
        - outlier-error
```
````

```bash
seldon pipeline load -f ./pipelines/error.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status error -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "error",
      "versions": [
        {
          "pipeline": {
            "name": "error",
            "uid": "cahee0nu7k0kl6oagsqg",
            "version": 1,
            "steps": [
              {
                "name": "outlier-error"
              }
            ],
            "output": {
              "steps": [
                "outlier-error.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-06-10T06:38:58.483729421Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer error --inference-mode grpc \
    '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```
````{collapse} Expand to see output
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
````

```bash
seldon pipeline infer error --inference-mode grpc \
    '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[100,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    Error: rpc error: code = FailedPrecondition desc = rpc error: code = Internal desc = Failed to process the request(s) for model instance 'outlier-error_1_0', message: TritonModelException: Outlier. Input sums to greater than 100
    
    At:
      /mnt/agent/models/outlier-error_1/1/model.py(108): execute
    
    Usage:
      seldon pipeline infer <pipelineName> (data) [flags]
    
    Flags:
      -f, --file-path string        inference payload file
          --header stringArray      add header key=value
      -h, --help                    help for infer
          --inference-host string   seldon inference host (default "0.0.0.0:9000")
          --inference-mode string   inference mode rest or grpc (default "rest")
      -i, --iterations int          inference iterations (default 1)
          --show-headers            show headers
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    rpc error: code = FailedPrecondition desc = rpc error: code = Internal desc = Failed to process the request(s) for model instance 'outlier-error_1_0', message: TritonModelException: Outlier. Input sums to greater than 100
    
    At:
      /mnt/agent/models/outlier-error_1/1/model.py(108): execute
    
```
````

```bash
seldon pipeline unload error
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload outlier-error
```
````{collapse} Expand to see output
```json

    {}
```
````
### Model Join with Trigger

Shows how steps in a pipeline can be triggered on data appearing. The trigger data is not sent to the step.


```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
cat ./models/tfsimple3.yaml
cat ./models/check.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple3
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: check
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/check"
      requirements:
      - triton
      - python
```
````

```bash
seldon model load -f ./models/tfsimple1.yaml 
seldon model load -f ./models/tfsimple2.yaml 
seldon model load -f ./models/tfsimple3.yaml 
seldon model load -f ./models/check.yaml 
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
```
````

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon model status tfsimple3 -w ModelAvailable | jq -M .
seldon model status check -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
```
````

```bash
cat ./pipelines/tfsimples-join-outlier.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: joincheck
      namespace: seldon-mesh
    spec:
      steps:
        - name: tfsimple1
        - name: tfsimple2
        - name: check
          inputs:
          - tfsimple1.outputs.OUTPUT0
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT
        - name: tfsimple3      
          inputs:
          - tfsimple1.outputs.OUTPUT0
          - tfsimple2.outputs.OUTPUT1
          tensorMap:
            tfsimple1.outputs.OUTPUT0: INPUT0
            tfsimple2.outputs.OUTPUT1: INPUT1
          triggers:
          - check.outputs.OUTPUT
      output:
        steps:
        - tfsimple3
```
````

```bash
seldon pipeline load -f ./pipelines/tfsimples-join-outlier.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status joincheck -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "joincheck",
      "versions": [
        {
          "pipeline": {
            "name": "joincheck",
            "uid": "caheevvu7k0kl6oagsr0",
            "version": 1,
            "steps": [
              {
                "name": "check",
                "inputs": [
                  "tfsimple1.outputs.OUTPUT0"
                ],
                "tensorMap": {
                  "tfsimple1.outputs.OUTPUT0": "INPUT"
                }
              },
              {
                "name": "tfsimple1"
              },
              {
                "name": "tfsimple2"
              },
              {
                "name": "tfsimple3",
                "inputs": [
                  "tfsimple1.outputs.OUTPUT0",
                  "tfsimple2.outputs.OUTPUT1"
                ],
                "tensorMap": {
                  "tfsimple1.outputs.OUTPUT0": "INPUT0",
                  "tfsimple2.outputs.OUTPUT1": "INPUT1"
                },
                "triggers": [
                  "check.outputs.OUTPUT"
                ]
              }
            ],
            "output": {
              "steps": [
                "tfsimple3.outputs"
              ]
            },
            "kubernetesMeta": {
              "namespace": "seldon-mesh"
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-06-10T06:41:04.702400893Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer joincheck --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "outputs": [
        {
          "name": "OUTPUT0",
          "datatype": "INT32",
          "shape": [
            "1",
            "16"
          ],
          "contents": {
            "intContents": [
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2
            ]
          }
        },
        {
          "name": "OUTPUT1",
          "datatype": "INT32",
          "shape": [
            "1",
            "16"
          ],
          "contents": {
            "intContents": [
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2,
              2
            ]
          }
        }
      ],
      "rawOutputContents": [
        "AgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAA==",
        "AgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAAIAAAACAAAAAgAAAA=="
      ]
    }
```
````

```bash
seldon pipeline infer joincheck --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

    Error: rpc error: code = FailedPrecondition desc = rpc error: code = Internal desc = Failed to process the request(s) for model instance 'check_1_0', message: TritonModelException: Outlier. Input sums to greater than 100
    
    At:
      /mnt/agent/models/check_1/1/model.py(107): execute
    
    Usage:
      seldon pipeline infer <pipelineName> (data) [flags]
    
    Flags:
      -f, --file-path string        inference payload file
          --header stringArray      add header key=value
      -h, --help                    help for infer
          --inference-host string   seldon inference host (default "0.0.0.0:9000")
          --inference-mode string   inference mode rest or grpc (default "rest")
      -i, --iterations int          inference iterations (default 1)
          --show-headers            show headers
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    parse error: Invalid numeric literal at line 1, column 4
```
````

```bash
seldon pipeline unload joincheck
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
seldon model unload tfsimple3
seldon model unload check
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
```
````
### Pipeline Input Tensors
Access to indivudal tensors in pipeline inputs


```bash
cat ./models/mul10.yaml
cat ./models/add10.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/mul10"
      requirements:
      - triton
      - python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton
      - python
```
````

```bash
seldon model load -f ./models/mul10.yaml 
seldon model load -f ./models/add10.yaml 
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
cat ./pipelines/pipeline-inputs.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: pipeline-inputs
      namespace: seldon-mesh
    spec:
      steps:
      - name: mul10
        inputs:
        - pipeline-inputs.inputs.INPUT0
        tensorMap:
          pipeline-inputs.inputs.INPUT0: INPUT
      - name: add10
        inputs:
        - pipeline-inputs.inputs.INPUT1
        tensorMap:
          pipeline-inputs.inputs.INPUT1: INPUT
      output:
        steps:
        - mul10
        - add10
    
```
````

```bash
seldon pipeline load -f ./pipelines/pipeline-inputs.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status pipeline-inputs -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "pipeline-inputs",
      "versions": [
        {
          "pipeline": {
            "name": "pipeline-inputs",
            "uid": "cahefk7u7k0kl6oagsrg",
            "version": 1,
            "steps": [
              {
                "name": "add10",
                "inputs": [
                  "pipeline-inputs.inputs.INPUT1"
                ],
                "tensorMap": {
                  "pipeline-inputs.inputs.INPUT1": "INPUT"
                }
              },
              {
                "name": "mul10",
                "inputs": [
                  "pipeline-inputs.inputs.INPUT0"
                ],
                "tensorMap": {
                  "pipeline-inputs.inputs.INPUT0": "INPUT"
                }
              }
            ],
            "output": {
              "steps": [
                "mul10.outputs",
                "add10.outputs"
              ]
            },
            "kubernetesMeta": {
              "namespace": "seldon-mesh"
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-06-10T06:42:24.461695292Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer pipeline-inputs --inference-mode grpc \
    '{"model_name":"pipeline","inputs":[{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```
````{collapse} Expand to see output
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
        },
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
        "AAAgQQAAoEEAAPBBAAAgQg==",
        "AAAwQQAAQEEAAFBBAABgQQ=="
      ]
    }
```
````

```bash
seldon pipeline unload pipeline-inputs
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload mul10
seldon model unload add10
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
### Trigger Joins

Shows how joins can be used for triggers as well.


```bash
cat ./models/mul10.yaml
cat ./models/add10.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/mul10"
      requirements:
      - triton
      - python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton
      - python
```
````

```bash
seldon model load -f ./models/mul10.yaml 
seldon model load -f ./models/add10.yaml 
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
cat ./pipelines/trigger-joins.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: trigger-joins
      namespace: seldon-mesh
    spec:
      steps:
      - name: mul10
        inputs:
        - trigger-joins.inputs.INPUT
        triggers:
        - trigger-joins.inputs.ok1
        - trigger-joins.inputs.ok2
        triggersJoinType: any
      - name: add10
        inputs:
        - trigger-joins.inputs.INPUT
        triggers:
        - trigger-joins.inputs.ok3
      output:
        steps:
        - mul10
        - add10
        stepsJoin: any
```
````

```bash
seldon pipeline load -f ./pipelines/trigger-joins.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status trigger-joins -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "trigger-joins",
      "versions": [
        {
          "pipeline": {
            "name": "trigger-joins",
            "uid": "cahefp7u7k0kl6oagss0",
            "version": 1,
            "steps": [
              {
                "name": "add10",
                "inputs": [
                  "trigger-joins.inputs.INPUT"
                ],
                "triggers": [
                  "trigger-joins.inputs.ok3"
                ]
              },
              {
                "name": "mul10",
                "inputs": [
                  "trigger-joins.inputs.INPUT"
                ],
                "triggers": [
                  "trigger-joins.inputs.ok1",
                  "trigger-joins.inputs.ok2"
                ],
                "triggersJoin": "ANY"
              }
            ],
            "output": {
              "steps": [
                "mul10.outputs",
                "add10.outputs"
              ],
              "stepsJoin": "ANY"
            },
            "kubernetesMeta": {
              "namespace": "seldon-mesh"
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-06-10T06:42:44.770994720Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer trigger-joins --inference-mode grpc \
    '{"model_name":"pipeline","inputs":[{"name":"ok1","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M . 
```
````{collapse} Expand to see output
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
````

```bash
seldon pipeline infer trigger-joins --inference-mode grpc \
    '{"model_name":"pipeline","inputs":[{"name":"ok3","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M . 
```
````{collapse} Expand to see output
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
````

```bash
seldon pipeline unload trigger-joins
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload mul10
seldon model unload add10
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```python

```
