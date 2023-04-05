## Seldon V2 Pipeline Examples

This notebook illustrates a series of Pipelines showing of different ways of combining flows of data and conditional logic. We assume you have Seldon Core V2 running locally.

### Models Used

 * `gs://seldon-models/triton/simple` an example Triton tensorflow model that takes 2 inputs INPUT0 and INPUT1 and adds them to produce OUTPUT0 and also subtracts INPUT1 from INPUT0 to produce OUTPUT1. See [here](https://github.com/triton-inference-server/server/tree/main/docs/examples/model_repository/simple) for the original source code and license.
 * Other models can be found at https://github.com/SeldonIO/triton-python-examples

### Model Chaining

Chain the output of one model into the next. Also shows chaning the tensor names via `tensorMap` to conform to the expected input tensor names of the second model.

```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple1
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple2
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki

```

```bash
seldon model load -f ./models/tfsimple1.yaml
seldon model load -f ./models/tfsimple2.yaml
```

```json
{}
{}

```

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

The pipeline below chains the output of `tfsimple1` into `tfsimple2`. As these models have compatible shape and data type this can be done. However, the output tensor names from `tfsimple1` need to be renamed to match the input tensor names for `tfsimple2`. We do this with the `tensorMap` feature.

The output of the Pipeline is the output from `tfsimple2`.

```bash
cat ./pipelines/tfsimples.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimples
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

```bash
seldon pipeline load -f ./pipelines/tfsimples.yaml
```

```json
{}

```

```bash
seldon pipeline status tfsimples -w PipelineReady| jq -M .
```

```json
{
  "pipelineName": "tfsimples",
  "versions": [
    {
      "pipeline": {
        "name": "tfsimples",
        "uid": "cgm2pdosogbs73emfvm0",
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
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-04-04T13:57:11.631385497Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer tfsimples \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
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
      ],
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    },
    {
      "data": [
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
      ],
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    }
  ]
}

```

```bash
seldon pipeline infer tfsimples --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
  ]
}

```

We use the Seldon CLI `pipeline inspect` feature to look at the data for all steps of the pipeline for the last data item passed through the pipeline (the default). This can be useful for debugging.

```bash
seldon pipeline inspect tfsimples
```

```
seldon.customer.default.model.tfsimple1.inputs	cg46198fh5ss73e09pm0	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
seldon.customer.default.model.tfsimple1.outputs	cg46198fh5ss73e09pm0	{"modelName":"tfsimple1_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
seldon.customer.default.model.tfsimple2.inputs	cg46198fh5ss73e09pm0	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
seldon.customer.default.model.tfsimple2.outputs	cg46198fh5ss73e09pm0	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}
seldon.customer.default.pipeline.tfsimples.inputs	cg46198fh5ss73e09pm0	{"modelName":"tfsimples","inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
seldon.customer.default.pipeline.tfsimples.outputs	cg46198fh5ss73e09pm0	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}

```

Next, we look get the output as json and use the `jq` tool to get just one value.

```bash
seldon pipeline inspect tfsimples --format json | jq -M .topics[0].msgs[0].value
```

```json
{
  "inputs": [
    {
      "name": "INPUT0",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          10,
          11,
          12,
          13,
          14,
          15,
          16
        ]
      }
    },
    {
      "name": "INPUT1",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          10,
          11,
          12,
          13,
          14,
          15,
          16
        ]
      }
    }
  ]
}

```

```bash
seldon pipeline unload tfsimples
```

```json
{}

```

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
```

```json
{}
{}

```

### Model Chaining from inputs

Chain the output of one model into the next. Shows using the input and outputs and combining.

```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple1
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple2
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki

```

```bash
seldon model load -f ./models/tfsimple1.yaml
seldon model load -f ./models/tfsimple2.yaml
```

```json
{}
{}

```

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

```bash
cat ./pipelines/tfsimples-input.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimples-input
spec:
  steps:
    - name: tfsimple1
    - name: tfsimple2
      inputs:
      - tfsimple1.inputs.INPUT0
      - tfsimple1.outputs.OUTPUT1
      tensorMap:
        tfsimple1.outputs.OUTPUT1: INPUT1
  output:
    steps:
    - tfsimple2

```

```bash
seldon pipeline load -f ./pipelines/tfsimples-input.yaml
```

```json
{}

```

```bash
seldon pipeline status tfsimples-input -w PipelineReady| jq -M .
```

```json
{
  "pipelineName": "tfsimples-input",
  "versions": [
    {
      "pipeline": {
        "name": "tfsimples-input",
        "uid": "cgm33165u83c73dgvr00",
        "version": 1,
        "steps": [
          {
            "name": "tfsimple1"
          },
          {
            "name": "tfsimple2",
            "inputs": [
              "tfsimple1.inputs.INPUT0",
              "tfsimple1.outputs.OUTPUT1"
            ],
            "tensorMap": {
              "tfsimple1.outputs.OUTPUT1": "INPUT1"
            }
          }
        ],
        "output": {
          "steps": [
            "tfsimple2.outputs"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-04-04T14:17:41.667004853Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer tfsimples-input \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9,
        10,
        11,
        12,
        13,
        14,
        15,
        16
      ],
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    },
    {
      "data": [
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9,
        10,
        11,
        12,
        13,
        14,
        15,
        16
      ],
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    }
  ]
}

```

```bash
seldon pipeline infer tfsimples-input --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
          1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          10,
          11,
          12,
          13,
          14,
          15,
          16
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
          1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          10,
          11,
          12,
          13,
          14,
          15,
          16
        ]
      }
    }
  ]
}

```

```bash
seldon pipeline unload tfsimples-input
```

```json
{}

```

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
```

```json
{}
{}

```

### Model Join

Join two flows of data from two models as input to a third model. This shows how individual flows of data can be combined.

```bash
cat ./models/tfsimple1.yaml
echo "---"
cat ./models/tfsimple2.yaml
echo "---"
cat ./models/tfsimple3.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple1
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple2
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple3
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki

```

```bash
seldon model load -f ./models/tfsimple1.yaml
seldon model load -f ./models/tfsimple2.yaml
seldon model load -f ./models/tfsimple3.yaml
```

```json
{}
{}
{}

```

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
seldon model status tfsimple3 -w ModelAvailable | jq -M .
```

```json
{}
{}
{}

```

In the pipeline below for the input to `tfsimple3` we join 1 output tensor each from the two previous models `tfsimple1` and `tfsimple2`. We need to use the `tensorMap` feature to rename each output tensor to one of the expected input tensors for the `tfsimple3` model.

```bash
cat ./pipelines/tfsimples-join.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: join
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

```bash
seldon pipeline load -f ./pipelines/tfsimples-join.yaml
```

```json
{}

```

```bash
seldon pipeline status join -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "join",
  "versions": [
    {
      "pipeline": {
        "name": "join",
        "uid": "cg3o00od8mqs73fqhh4g",
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
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-07T18:18:43.910647312Z",
        "modelsReady": true
      }
    }
  ]
}

```

The outputs are the sequence "2,4,6..." which conforms to the logic of this model (addition and subtraction) when fed the output of the first two models.

```bash
seldon pipeline infer join --inference-mode grpc \
    '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
  ]
}

```

```bash
seldon pipeline unload join
```

```json
{}

```

```bash
seldon model unload tfsimple1
seldon model unload tfsimple2
seldon model unload tfsimple3
```

```json
{}
{}
{}

```

### Conditional

Shows conditional data flows - one of two models is run based on output tensors from first.

```bash
cat ./models/conditional.yaml
echo "---"
cat ./models/add10.yaml
echo "---"
cat ./models/mul10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: conditional
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/conditional"
  requirements:
  - triton
  - python
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/add10"
  requirements:
  - triton
  - python
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/mul10"
  requirements:
  - triton
  - python

```

```bash
seldon model load -f ./models/conditional.yaml
seldon model load -f ./models/add10.yaml
seldon model load -f ./models/mul10.yaml
```

```json
{}
{}
{}

```

```bash
seldon model status conditional -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
seldon model status mul10 -w ModelAvailable | jq -M .
```

```json
{}
{}
{}

```

Here we assume the `conditional` model can output two tensors OUTPUT0 and OUTPUT1 but only outputs the former if the CHOICE input tensor is set to 0 otherwise it outputs tensor OUTPUT1. By this means only one of the two downstream models will receive data and run. The `output` steps does an `any` join from both models and whichever data appears first will be sent as output to pipeline. As in this case only 1 of the two models `add10` and `mul10` runs we will receive their output.

```bash
cat ./pipelines/conditional.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimple-conditional
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

```bash
seldon pipeline load -f ./pipelines/conditional.yaml
```

```json
{}

```

```bash
seldon pipeline status tfsimple-conditional -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "tfsimple-conditional",
  "versions": [
    {
      "pipeline": {
        "name": "tfsimple-conditional",
        "uid": "cg3o05od8mqs73fqhh50",
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
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-07T18:19:04.254724839Z",
        "modelsReady": true
      }
    }
  ]
}

```

The `mul10` model will run as the CHOICE tensor is set to 0.

```bash
seldon pipeline infer tfsimple-conditional --inference-mode grpc \
 '{"model_name":"conditional","inputs":[{"name":"CHOICE","contents":{"int_contents":[0]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
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

The `add10` model will run as the CHOICE tensor is not set to zero.

```bash
seldon pipeline infer tfsimple-conditional --inference-mode grpc \
 '{"model_name":"conditional","inputs":[{"name":"CHOICE","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
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
seldon pipeline unload tfsimple-conditional
```

```json
{}

```

```bash
seldon model unload conditional
seldon model unload add10
seldon model unload mul10
```

```json
{}
{}
{}

```

### Pipeline Input Tensors
Access to indivudal tensors in pipeline inputs

```bash
cat ./models/mul10.yaml
echo "---"
cat ./models/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/mul10"
  requirements:
  - triton
  - python
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/add10"
  requirements:
  - triton
  - python

```

```bash
seldon model load -f ./models/mul10.yaml
seldon model load -f ./models/add10.yaml
```

```json
{}
{}

```

```bash
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

This pipeline shows how we can access pipeline inputs INPUT0 and INPUT1 from different steps.

```bash
cat ./pipelines/pipeline-inputs.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-inputs
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

```bash
seldon pipeline load -f ./pipelines/pipeline-inputs.yaml
```

```json
{}

```

```bash
seldon pipeline status pipeline-inputs -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "pipeline-inputs",
  "versions": [
    {
      "pipeline": {
        "name": "pipeline-inputs",
        "uid": "cg3o0agd8mqs73fqhh5g",
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
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-07T18:19:22.326945896Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer pipeline-inputs --inference-mode grpc \
    '{"model_name":"pipeline","inputs":[{"name":"INPUT0","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]},{"name":"INPUT1","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
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
  ]
}

```

```bash
seldon pipeline unload pipeline-inputs
```

```json
{}

```

```bash
seldon model unload mul10
seldon model unload add10
```

```json
{}
{}

```

### Trigger Joins

Shows how joins can be used for triggers as well.

```bash
cat ./models/mul10.yaml
cat ./models/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/mul10"
  requirements:
  - triton
  - python
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_22-11/add10"
  requirements:
  - triton
  - python

```

```bash
seldon model load -f ./models/mul10.yaml
seldon model load -f ./models/add10.yaml
```

```json
{}
{}

```

```bash
seldon model status mul10 -w ModelAvailable | jq -M .
seldon model status add10 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

Here we required tensors names `ok1` or `ok2` to exist on pipeline inputs to run the `mul10` model but require tensor `ok3` to exist on pipeline inputs to run the `add10` model. The logic on `mul10` is handled by a trigger join of `any` meaning either of these input data can exist to satisfy the trigger join.

```bash
cat ./pipelines/trigger-joins.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: trigger-joins
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

```bash
seldon pipeline load -f ./pipelines/trigger-joins.yaml
```

```json
{}

```

```bash
seldon pipeline status trigger-joins -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "trigger-joins",
  "versions": [
    {
      "pipeline": {
        "name": "trigger-joins",
        "uid": "cg3o0f0d8mqs73fqhh60",
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
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-07T18:19:40.673234267Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer trigger-joins --inference-mode grpc \
    '{"model_name":"pipeline","inputs":[{"name":"ok1","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
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
seldon pipeline infer trigger-joins --inference-mode grpc \
    '{"model_name":"pipeline","inputs":[{"name":"ok3","contents":{"fp32_contents":[1]},"datatype":"FP32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
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
seldon pipeline unload trigger-joins
```

```json
{}

```

```bash
seldon model unload mul10
seldon model unload add10
```

```json
{}
{}

```

```python

```
