# Seldon V2 Non Kubernetes Pipeline Examples


 * Build if needed and place `seldon` binary in your path
    * run `make build-seldon` from operator folder and add bin folder to PATH
 * Run Seldon V2 `make deploy-local` from top level folder


```bash
which seldon
```

    /home/clive/work/scv2/seldon-core-v2/operator/bin/seldon
```
````
## Model Chaining


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
seldon model status --model-name tfsimple1 -w ModelAvailable | jq -M .
seldon model status --model-name tfsimple2 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "tfsimple1",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:20:20.260330848Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:20:20.260330848Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple1",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "tfsimple2",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:20:20.361156338Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:20:20.361156338Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple2",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
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
seldon pipeline status -p tfsimples -w PipelineReady| jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "tfsimples",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimples",
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
            "lastChangeTimestamp": "2022-04-26T10:20:56.271848455Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer -p tfsimples \
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
seldon pipeline infer -p tfsimples --inference-mode grpc \
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
seldon pipeline unload -p tfsimples
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload --model-name tfsimple1
seldon model unload --model-name tfsimple2
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
## Model Join


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
seldon model status --model-name tfsimple1 -w ModelAvailable | jq -M .
seldon model status --model-name tfsimple2 -w ModelAvailable | jq -M .
seldon model status --model-name tfsimple3 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "tfsimple1",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:21:11.274867202Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:21:11.274867202Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple1",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "tfsimple2",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:21:11.425542338Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:21:11.425542338Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple2",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "tfsimple3",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:21:11.571780123Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:21:11.571780123Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple3",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
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
seldon pipeline status -p join -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "join",
      "versions": [
        {
          "pipeline": {
            "name": "join",
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
            "lastChangeTimestamp": "2022-04-26T10:21:43.374763891Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer -p join --inference-mode grpc \
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
seldon pipeline unload -p join
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload --model-name tfsimple1
seldon model unload --model-name tfsimple2
seldon model unload --model-name tfsimple3
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````
## Conditional


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
      - triton-python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton-python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/mul10"
      requirements:
      - triton-python
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
seldon model status --model-name conditional -w ModelAvailable | jq -M .
seldon model status --model-name add10 -w ModelAvailable | jq -M .
seldon model status --model-name mul10 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "conditional",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:22:04.449175941Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:22:04.449175941Z"
          },
          "modelDefn": {
            "meta": {
              "name": "conditional",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/conditional",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "add10",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:22:04.721024552Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:22:04.721024552Z"
          },
          "modelDefn": {
            "meta": {
              "name": "add10",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/add10",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "mul10",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:22:05.002034725Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:22:05.002034725Z"
          },
          "modelDefn": {
            "meta": {
              "name": "mul10",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/mul10",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
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
seldon pipeline status -p tfsimple-conditional -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "tfsimple-conditional",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple-conditional",
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
            "lastChangeTimestamp": "2022-04-26T10:22:47.702756329Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer -p tfsimple-conditional --inference-mode grpc \
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
seldon pipeline infer -p tfsimple-conditional --inference-mode grpc \
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
seldon pipeline unload -p tfsimple-conditional
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload --model-name conditional
seldon model unload --model-name add10
seldon model unload --model-name mul10
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````
## Error
An example which errors is arguments sum to greater than 100


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
      - triton-python
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
seldon model status --model-name outlier-error -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "outlier-error",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:24:59.106635468Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:24:59.106635468Z"
          },
          "modelDefn": {
            "meta": {
              "name": "outlier-error",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/outlier",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
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
seldon pipeline status -p error -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "error",
      "versions": [
        {
          "pipeline": {
            "name": "error",
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
            "lastChangeTimestamp": "2022-04-26T10:25:28.833606915Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer -p error --inference-mode grpc \
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
seldon pipeline infer -p error --inference-mode grpc \
    '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[100,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    Error: rpc error: code = FailedPrecondition desc = rpc error: code = Internal desc = Failed to process the request(s) for model instance 'outlier-error_1_0', message: TritonModelException: Outlier. Input sums to greater than 100
    
    At:
      /mnt/agent/models/outlier-error_1/1/model.py(108): execute
    
    Usage:
      seldon pipeline infer [flags]
    
    Flags:
      -f, --file-path string        inference payload file
      -h, --help                    help for infer
          --inference-host string   seldon inference host (default "0.0.0.0")
          --inference-mode string   inference mode rest or grpc (default "rest")
          --inference-port int      seldon scheduler port (default 9000)
      -i, --iterations int          inference iterations (default 1)
      -p, --pipeline-name string    pipeline name for inference
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    rpc error: code = FailedPrecondition desc = rpc error: code = Internal desc = Failed to process the request(s) for model instance 'outlier-error_1_0', message: TritonModelException: Outlier. Input sums to greater than 100
    
    At:
      /mnt/agent/models/outlier-error_1/1/model.py(108): execute
    
```
````

```bash
seldon pipeline unload -p error
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload --model-name outlier-error
```
````{collapse} Expand to see output
```json

    {}
```
````
## Outlier
An example runs only if no outlier


```bash
cat ./models/outlier-error.yaml
cat ./models/add10.yaml
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
      - triton-python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton-python
```
````

```bash
seldon model load -f ./models/outlier-error.yaml 
seldon model load -f ./models/add10.yaml 
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
seldon model status --model-name outlier-error -w ModelAvailable | jq -M .
seldon model status --model-name add10 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "outlier-error",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:26:15.119768627Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:26:15.119768627Z"
          },
          "modelDefn": {
            "meta": {
              "name": "outlier-error",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/outlier",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "add10",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:26:15.391611238Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:26:15.391611238Z"
          },
          "modelDefn": {
            "meta": {
              "name": "add10",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/add10",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
```
````

```bash
cat ./pipelines/outlier.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: outlier
    spec:
      steps:
        - name: outlier-error
        - name: add10
          triggers:
          - outlier-error
      output:
        steps:
        - add10
```
````

```bash
seldon pipeline load -f ./pipelines/outlier.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status -p outlier -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "outlier",
      "versions": [
        {
          "pipeline": {
            "name": "outlier",
            "steps": [
              {
                "name": "add10",
                "triggers": [
                  "outlier-error.outputs"
                ]
              },
              {
                "name": "outlier-error"
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
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-04-26T10:26:54.769531603Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer -p outlier --inference-mode grpc \
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
seldon pipeline infer -p outlier --inference-mode grpc \
    '{"model_name":"outlier","inputs":[{"name":"INPUT","contents":{"fp32_contents":[100,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq .
```
````{collapse} Expand to see output
```json

    Error: rpc error: code = FailedPrecondition desc = rpc error: code = Internal desc = Failed to process the request(s) for model instance 'outlier-error_1_0', message: TritonModelException: Outlier. Input sums to greater than 100
    
    At:
      /mnt/agent/models/outlier-error_1/1/model.py(108): execute
    
    Usage:
      seldon pipeline infer [flags]
    
    Flags:
      -f, --file-path string        inference payload file
      -h, --help                    help for infer
          --inference-host string   seldon inference host (default "0.0.0.0")
          --inference-mode string   inference mode rest or grpc (default "rest")
          --inference-port int      seldon scheduler port (default 9000)
      -i, --iterations int          inference iterations (default 1)
      -p, --pipeline-name string    pipeline name for inference
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    parse error: Invalid numeric literal at line 1, column 4
```
````

```bash
seldon pipeline unload -p outlier
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload --model-name outlier-error
seldon model unload --model-name add10
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
## Model Join with Trigger


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
      - triton-python
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
seldon model status --model-name tfsimple1 -w ModelAvailable | jq -M .
seldon model status --model-name tfsimple2 -w ModelAvailable | jq -M .
seldon model status --model-name tfsimple3 -w ModelAvailable | jq -M .
seldon model status --model-name check -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "tfsimple1",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:29:43.097262399Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:29:43.097262399Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple1",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "tfsimple2",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:29:43.199880236Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:29:43.199880236Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple2",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "tfsimple3",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:29:43.302137780Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:29:43.302137780Z"
          },
          "modelDefn": {
            "meta": {
              "name": "tfsimple3",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/simple",
              "requirements": [
                "tensorflow"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "check",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:29:43.739386729Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:29:43.739386729Z"
          },
          "modelDefn": {
            "meta": {
              "name": "check",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/check",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
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
seldon pipeline status -p joincheck -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "joincheck",
      "versions": [
        {
          "pipeline": {
            "name": "joincheck",
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
            "lastChangeTimestamp": "2022-04-26T10:30:34.333430917Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer -p joincheck --inference-mode grpc \
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
seldon pipeline infer -p joincheck --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

    Error: rpc error: code = FailedPrecondition desc = rpc error: code = Internal desc = Failed to process the request(s) for model instance 'check_1_0', message: TritonModelException: Outlier. Input sums to greater than 100
    
    At:
      /mnt/agent/models/check_1/1/model.py(107): execute
    
    Usage:
      seldon pipeline infer [flags]
    
    Flags:
      -f, --file-path string        inference payload file
      -h, --help                    help for infer
          --inference-host string   seldon inference host (default "0.0.0.0")
          --inference-mode string   inference mode rest or grpc (default "rest")
          --inference-port int      seldon scheduler port (default 9000)
      -i, --iterations int          inference iterations (default 1)
      -p, --pipeline-name string    pipeline name for inference
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    parse error: Invalid numeric literal at line 1, column 4
```
````

```bash
seldon pipeline unload -p joincheck
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload --model-name tfsimple1
seldon model unload --model-name tfsimple2
seldon model unload --model-name tfsimple3
seldon model unload --model-name check
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
```
````
## Pipeline Input Tensors
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
      - triton-python
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton-python
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
seldon model status --model-name mul10 -w ModelAvailable | jq -M .
seldon model status --model-name add10 -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "mul10",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:31:13.008913813Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:31:13.008913813Z"
          },
          "modelDefn": {
            "meta": {
              "name": "mul10",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/mul10",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "add10",
      "versions": [
        {
          "version": 1,
          "serverName": "triton",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-04-26T10:31:12.726176321Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-04-26T10:31:12.726176321Z"
          },
          "modelDefn": {
            "meta": {
              "name": "add10",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/triton/add10",
              "requirements": [
                "triton-python"
              ]
            },
            "deploymentSpec": {
              "replicas": 1,
              "minReplicas": 1
            }
          }
        }
      ]
    }
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
seldon pipeline status -p pipeline-inputs -w PipelineReady | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "pipeline-inputs",
      "versions": [
        {
          "pipeline": {
            "name": "pipeline-inputs",
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
            "lastChangeTimestamp": "2022-04-26T10:31:46.743076681Z"
          }
        }
      ]
    }
```
````

```bash
seldon pipeline infer -p pipeline-inputs --inference-mode grpc \
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
seldon pipeline unload -p pipeline-inputs
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload --model-name mul10
seldon model unload --model-name add10
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```python

```
