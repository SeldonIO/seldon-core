## Seldon V2 Pipeline to Pipeline Examples

This notebook illustrates a series of Pipelines that are joined together.

### Models Used

 * `gs://seldon-models/triton/simple` an example Triton tensorflow model that takes 2 inputs INPUT0 and INPUT1 and adds them to produce OUTPUT0 and also subtracts INPUT1 from INPUT0 to produce OUTPUT1. See [here](https://github.com/triton-inference-server/server/tree/main/docs/examples/model_repository/simple) for the original source code and license.
 * Other models can be found at https://github.com/SeldonIO/triton-python-examples

### Pipeline pulling from one other Pipeline

![pipeline-to-pipeline](img_pipeline1.jpg)



```python
!cat ./models/tfsimple1.yaml
!cat ./models/tfsimple2.yaml
```

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



```python
!seldon model load -f ./models/tfsimple1.yaml 
!seldon model load -f ./models/tfsimple2.yaml 
```

    {}
    {}



```python
!seldon model status tfsimple1 -w ModelAvailable | jq -M .
!seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!cat ./pipelines/tfsimple.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple
    spec:
      steps:
        - name: tfsimple1
      output:
        steps:
        - tfsimple1



```python
!seldon pipeline load -f ./pipelines/tfsimple.yaml
```

    {}



```python
!seldon pipeline status tfsimple -w PipelineReady| jq -M .
```

    {
      "pipelineName": "tfsimple",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple",
            "uid": "ceevilesc0ns73er67kg",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple1"
              }
            ],
            "output": {
              "steps": [
                "tfsimple1.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-12-17T17:16:05.915396256Z",
            "modelsReady": true
          }
        }
      ]
    }



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
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



```python
!cat ./pipelines/tfsimple-extended.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended
    spec:
      input:
        externalInputs:
          - tfsimple.outputs
        tensorMap:
          tfsimple.outputs.OUTPUT0: INPUT0
          tfsimple.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2



```python
!seldon pipeline load -f ./pipelines/tfsimple-extended.yaml
```

    {}



```python
!seldon pipeline status tfsimple-extended -w PipelineReady| jq -M .
```

    {
      "pipelineName": "tfsimple-extended",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple-extended",
            "uid": "ceevin6sc0ns73er67l0",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple2"
              }
            ],
            "output": {
              "steps": [
                "tfsimple2.outputs"
              ]
            },
            "kubernetesMeta": {},
            "input": {
              "externalInputs": [
                "tfsimple.outputs"
              ],
              "tensorMap": {
                "tfsimple.outputs.OUTPUT0": "INPUT0",
                "tfsimple.outputs.OUTPUT1": "INPUT1"
              }
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-12-17T17:16:12.837674035Z",
            "modelsReady": true
          }
        }
      ]
    }



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' 
```

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
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0
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



```python
!seldon pipeline inspect tfsimple
```

    seldon.default.model.tfsimple1.inputs	ceevins7aa7c73d9cfo0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.model.tfsimple1.outputs	ceevins7aa7c73d9cfo0	{"modelName":"tfsimple1_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple.inputs	ceevins7aa7c73d9cfo0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.pipeline.tfsimple.outputs	ceevins7aa7c73d9cfo0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}



```python
!seldon pipeline inspect tfsimple-extended
```

    seldon.default.model.tfsimple2.inputs	ceevins7aa7c73d9cfo0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceevins7aa7c73d9cfo0	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}
    seldon.default.pipeline.tfsimple-extended.inputs	ceevins7aa7c73d9cfo0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended.outputs	ceevins7aa7c73d9cfo0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline unload tfsimple-extended
!seldon pipeline unload tfsimple
```

    {}
    {}



```python
!seldon model unload tfsimple1
!seldon model unload tfsimple2
```

    {}
    {}


### Pipeline pulling from two other Pipelines

![pipeline-to-pipeline](img_pipeline2.jpg)



```python
!cat ./models/tfsimple1.yaml
!cat ./models/tfsimple2.yaml
```

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



```python
!seldon model load -f ./models/tfsimple1.yaml 
!seldon model load -f ./models/tfsimple2.yaml 
```

    {}
    {}



```python
!seldon model status tfsimple1 -w ModelAvailable | jq -M .
!seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!cat ./pipelines/tfsimple.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple
    spec:
      steps:
        - name: tfsimple1
      output:
        steps:
        - tfsimple1



```python
!seldon pipeline load -f ./pipelines/tfsimple.yaml
```

    {}



```python
!seldon pipeline status tfsimple -w PipelineReady| jq -M .
```

    {
      "pipelineName": "tfsimple",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple",
            "uid": "ceevj2msc0ns73er67lg",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple1"
              }
            ],
            "output": {
              "steps": [
                "tfsimple1.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-12-17T17:16:59.017037835Z",
            "modelsReady": true
          }
        }
      ]
    }



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
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



```python
!cat ./pipelines/tfsimple-extended.yaml
!echo "---"
!cat ./pipelines/tfsimple-extended2.yaml
!echo "---"
!cat ./pipelines/tfsimple-combined.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended
    spec:
      input:
        externalInputs:
          - tfsimple.outputs
        tensorMap:
          tfsimple.outputs.OUTPUT0: INPUT0
          tfsimple.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended2
    spec:
      input:
        externalInputs:
          - tfsimple.outputs
        tensorMap:
          tfsimple.outputs.OUTPUT0: INPUT0
          tfsimple.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-combined
    spec:
      input:
        externalInputs:
          - tfsimple-extended.outputs.OUTPUT0
          - tfsimple-extended2.outputs.OUTPUT1
        tensorMap:
          tfsimple-extended.outputs.OUTPUT0: INPUT0
          tfsimple-extended2.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2



```python
!seldon pipeline load -f ./pipelines/tfsimple-extended.yaml
!seldon pipeline load -f ./pipelines/tfsimple-extended2.yaml
!seldon pipeline load -f ./pipelines/tfsimple-combined.yaml
```

    {}
    {}
    {}



```python
!seldon pipeline status tfsimple-extended -w PipelineReady
!seldon pipeline status tfsimple-extended2 -w PipelineReady
!seldon pipeline status tfsimple-combined -w PipelineReady
```

    {"pipelineName":"tfsimple-extended", "versions":[{"pipeline":{"name":"tfsimple-extended", "uid":"ceevj3usc0ns73er67m0", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple.outputs"], "tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0", "tfsimple.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:17:04.079245812Z", "modelsReady":true}}]}
    {"pipelineName":"tfsimple-extended2", "versions":[{"pipeline":{"name":"tfsimple-extended2", "uid":"ceevj46sc0ns73er67mg", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple.outputs"], "tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0", "tfsimple.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:17:04.226769422Z", "modelsReady":true}}]}
    {"pipelineName":"tfsimple-combined", "versions":[{"pipeline":{"name":"tfsimple-combined", "uid":"ceevj46sc0ns73er67n0", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple-extended.outputs.OUTPUT0", "tfsimple-extended2.outputs.OUTPUT1"], "tensorMap":{"tfsimple-extended.outputs.OUTPUT0":"INPUT0", "tfsimple-extended2.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:17:04.466100483Z", "modelsReady":true}}]}



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' 
```

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
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0
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



```python
!seldon pipeline inspect tfsimple
```

    seldon.default.model.tfsimple1.inputs	ceevj4k7aa7c73d9cfp0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.model.tfsimple1.outputs	ceevj4k7aa7c73d9cfp0	{"modelName":"tfsimple1_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple.inputs	ceevj4k7aa7c73d9cfp0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.pipeline.tfsimple.outputs	ceevj4k7aa7c73d9cfp0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}



```python
!seldon pipeline inspect tfsimple-extended
```

    seldon.default.model.tfsimple2.inputs	ceevj347aa7c73d9cfog	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceepua47aa7c73d9cfj0	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-extended.inputs	ceevj4k7aa7c73d9cfp0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended.outputs	ceevj4k7aa7c73d9cfp0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline inspect tfsimple-extended2
```

    seldon.default.model.tfsimple2.inputs	ceevj4k7aa7c73d9cfp0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceeq4ls7aa7c73d9cfk0	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-extended2.inputs	ceevj4k7aa7c73d9cfp0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended2.outputs	ceevj4k7aa7c73d9cfp0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline inspect tfsimple-combined
```

    seldon.default.model.tfsimple2.inputs	ceevj4k7aa7c73d9cfp0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceespfk7aa7c73d9cfmg	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-combined.inputs	ceevj4k7aa7c73d9cfp0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.pipeline.tfsimple-combined.outputs	ceespfk7aa7c73d9cfmg	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}



```python
!seldon pipeline unload tfsimple-extended
!seldon pipeline unload tfsimple-extended2
!seldon pipeline unload tfsimple-combined
!seldon pipeline unload tfsimple
```

    {}
    {}
    {}
    {}



```python
!seldon model unload tfsimple1
!seldon model unload tfsimple2
```

    {}
    {}


### Pipeline pullin from one pipeline with a trigger to another

![pipeline-to-pipeline](img_pipeline3.jpg)



```python
!cat ./models/tfsimple1.yaml
!cat ./models/tfsimple2.yaml
```

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



```python
!seldon model load -f ./models/tfsimple1.yaml 
!seldon model load -f ./models/tfsimple2.yaml 
```

    {}
    {}



```python
!seldon model status tfsimple1 -w ModelAvailable | jq -M .
!seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!cat ./pipelines/tfsimple.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple
    spec:
      steps:
        - name: tfsimple1
      output:
        steps:
        - tfsimple1



```python
!seldon pipeline load -f ./pipelines/tfsimple.yaml
```

    {}



```python
!seldon pipeline status tfsimple -w PipelineReady| jq -M .
```

    {
      "pipelineName": "tfsimple",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple",
            "uid": "ceevja6sc0ns73er67ng",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple1"
              }
            ],
            "output": {
              "steps": [
                "tfsimple1.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-12-17T17:17:29.011892538Z",
            "modelsReady": true
          }
        }
      ]
    }



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
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



```python
!cat ./pipelines/tfsimple-extended.yaml
!echo "---"
!cat ./pipelines/tfsimple-extended2.yaml
!echo "---"
!cat ./pipelines/tfsimple-combined-trigger.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended
    spec:
      input:
        externalInputs:
          - tfsimple.outputs
        tensorMap:
          tfsimple.outputs.OUTPUT0: INPUT0
          tfsimple.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended2
    spec:
      input:
        externalInputs:
          - tfsimple.outputs
        tensorMap:
          tfsimple.outputs.OUTPUT0: INPUT0
          tfsimple.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-combined-trigger
    spec:
      input:
        externalInputs:
          - tfsimple-extended.outputs
        externalTriggers:
          - tfsimple-extended2.outputs
        tensorMap:
          tfsimple-extended.outputs.OUTPUT0: INPUT0
          tfsimple-extended.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2



```python
!seldon pipeline load -f ./pipelines/tfsimple-extended.yaml
!seldon pipeline load -f ./pipelines/tfsimple-extended2.yaml
!seldon pipeline load -f ./pipelines/tfsimple-combined-trigger.yaml
```

    {}
    {}
    {}



```python
!seldon pipeline status tfsimple-extended -w PipelineReady
!seldon pipeline status tfsimple-extended2 -w PipelineReady
!seldon pipeline status tfsimple-combined-trigger -w PipelineReady
```

    {"pipelineName":"tfsimple-extended", "versions":[{"pipeline":{"name":"tfsimple-extended", "uid":"ceevjbmsc0ns73er67o0", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple.outputs"], "tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0", "tfsimple.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:17:34.180121811Z", "modelsReady":true}}]}
    {"pipelineName":"tfsimple-extended2", "versions":[{"pipeline":{"name":"tfsimple-extended2", "uid":"ceevjbmsc0ns73er67og", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple.outputs"], "tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0", "tfsimple.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:17:34.307458250Z", "modelsReady":true}}]}
    {"pipelineName":"tfsimple-combined-trigger", "versions":[{"pipeline":{"name":"tfsimple-combined-trigger", "uid":"ceevjbmsc0ns73er67p0", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple-extended.outputs"], "externalTriggers":["tfsimple-extended2.outputs"], "tensorMap":{"tfsimple-extended.outputs.OUTPUT0":"INPUT0", "tfsimple-extended.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:17:34.550045739Z", "modelsReady":true}}]}



```python
!seldon pipeline infer tfsimple --header x-request-id=test-id3 \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' 
```

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
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0
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



```python
!seldon pipeline inspect tfsimple
```

    seldon.default.model.tfsimple1.inputs	test-id3	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.model.tfsimple1.outputs	test-id3	{"modelName":"tfsimple1_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple.inputs	test-id3	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.pipeline.tfsimple.outputs	test-id3	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}



```python
!seldon pipeline inspect tfsimple-extended
```

    seldon.default.model.tfsimple2.inputs	ceeq4ok7aa7c73d9cfkg	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceepua47aa7c73d9cfj0	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-extended.inputs	test-id3	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended.outputs	test-id3	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline inspect tfsimple-extended2
```

    seldon.default.model.tfsimple2.inputs	ceeq4ok7aa7c73d9cfkg	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceeq4ok7aa7c73d9cfkg	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-extended2.inputs	test-id3	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended2.outputs	test-id3	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline inspect tfsimple-combined-trigger
```

    seldon.default.model.tfsimple2.inputs	ceepua47aa7c73d9cfj0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceeq4ls7aa7c73d9cfk0	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-combined-trigger.inputs	test-id3	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.pipeline.tfsimple-combined-trigger.outputs	ceespfk7aa7c73d9cfmg	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}



```python
!seldon pipeline unload tfsimple-extended
!seldon pipeline unload tfsimple-extended2
!seldon pipeline unload tfsimple-combined-trigger
!seldon pipeline unload tfsimple
```

    {}
    {}
    {}
    {}



```python
!seldon model unload tfsimple1
!seldon model unload tfsimple2
```

    {}
    {}


### Pipeline pulling from one other Pipeline Step

![pipeline-to-pipeline](img_pipeline4.jpg)



```python
!cat ./models/tfsimple1.yaml
!cat ./models/tfsimple2.yaml
```

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



```python
!seldon model load -f ./models/tfsimple1.yaml 
!seldon model load -f ./models/tfsimple2.yaml 
```

    {}
    {}



```python
!seldon model status tfsimple1 -w ModelAvailable | jq -M .
!seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!cat ./pipelines/tfsimple.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple
    spec:
      steps:
        - name: tfsimple1
      output:
        steps:
        - tfsimple1



```python
!seldon pipeline load -f ./pipelines/tfsimple.yaml
```

    {}



```python
!seldon pipeline status tfsimple -w PipelineReady| jq -M .
```

    {
      "pipelineName": "tfsimple",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple",
            "uid": "ceevjh6sc0ns73er67pg",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple1"
              }
            ],
            "output": {
              "steps": [
                "tfsimple1.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-12-17T17:17:56.809691365Z",
            "modelsReady": true
          }
        }
      ]
    }



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
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



```python
!cat ./pipelines/tfsimple-extended-step.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended-step
    spec:
      input:
        externalInputs:
          - tfsimple.step.tfsimple1.outputs
        tensorMap:
          tfsimple.step.tfsimple1.outputs.OUTPUT0: INPUT0
          tfsimple.step.tfsimple1.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2



```python
!seldon pipeline load -f ./pipelines/tfsimple-extended-step.yaml
```

    {}



```python
!seldon pipeline status tfsimple-extended-step -w PipelineReady| jq -M .
```

    {
      "pipelineName": "tfsimple-extended-step",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple-extended-step",
            "uid": "ceevkf6sc0ns73er67q0",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple2"
              }
            ],
            "output": {
              "steps": [
                "tfsimple2.outputs"
              ]
            },
            "kubernetesMeta": {},
            "input": {
              "externalInputs": [
                "tfsimple.step.tfsimple1.outputs"
              ],
              "tensorMap": {
                "tfsimple.step.tfsimple1.outputs.OUTPUT0": "INPUT0",
                "tfsimple.step.tfsimple1.outputs.OUTPUT1": "INPUT1"
              }
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-12-17T17:19:56.669740116Z",
            "modelsReady": true
          }
        }
      ]
    }



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' 
```

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
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0
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



```python
!seldon pipeline inspect tfsimple --verbose
```

    seldon.default.model.tfsimple1.inputs	ceevkfs7aa7c73d9cfs0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}		pipeline=[tfsimple]	traceparent=[00-01d19de78cff2a815a8f20ba38f7a009-c18e60cf70b2c9fa-01]	x-forwarded-proto=[http]	x-envoy-expected-rq-timeout-ms=[60000]
    seldon.default.model.tfsimple1.outputs	ceevkfs7aa7c73d9cfs0	{"modelName":"tfsimple1_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}		pipeline=[tfsimple]	x-envoy-upstream-service-time=[0]	x-seldon-route=[:tfsimple1_1:]	x-request-id=[ceevkfr9me5c73c9i7jg]	traceparent=[00-01d19de78cff2a815a8f20ba38f7a009-cafaccc7a3930447-01]	x-forwarded-proto=[http]	x-envoy-expected-rq-timeout-ms=[60000]
    seldon.default.pipeline.tfsimple.inputs	ceevkfs7aa7c73d9cfs0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}		x-forwarded-proto=[http]	x-envoy-expected-rq-timeout-ms=[60000]	pipeline=[tfsimple]	traceparent=[00-01d19de78cff2a815a8f20ba38f7a009-03db08cfee014b2c-01]
    seldon.default.pipeline.tfsimple.outputs	ceevkfs7aa7c73d9cfs0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}		x-envoy-upstream-service-time=[0]	x-seldon-route=[:tfsimple1_1:]	x-request-id=[ceevkfr9me5c73c9i7jg]	pipeline=[tfsimple]	traceparent=[00-01d19de78cff2a815a8f20ba38f7a009-b4a87cbb29038c9c-01]	x-forwarded-proto=[http]	x-envoy-expected-rq-timeout-ms=[60000]



```python
!seldon pipeline inspect tfsimple-extended-step
```

    seldon.default.model.tfsimple2.inputs	ceevkfs7aa7c73d9cfs0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceepucc7aa7c73d9cfjg	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-extended-step.inputs	ceevkfs7aa7c73d9cfs0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended-step.outputs	ceeq5ss7aa7c73d9cfm0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline unload tfsimple-extended-step
!seldon pipeline unload tfsimple
```

    {}
    {}



```python
!seldon model unload tfsimple1
!seldon model unload tfsimple2
```

    {}
    {}


### Pipeline pulling from two other Pipeline steps from same model

![pipeline-to-pipeline](img_pipeline5.jpg)



```python
!cat ./models/tfsimple1.yaml
!cat ./models/tfsimple2.yaml
```

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



```python
!seldon model load -f ./models/tfsimple1.yaml 
!seldon model load -f ./models/tfsimple2.yaml 
```

    {}
    {}



```python
!seldon model status tfsimple1 -w ModelAvailable | jq -M .
!seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

    {}
    {}



```python
!cat ./pipelines/tfsimple.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple
    spec:
      steps:
        - name: tfsimple1
      output:
        steps:
        - tfsimple1



```python
!seldon pipeline load -f ./pipelines/tfsimple.yaml
```

    {}



```python
!seldon pipeline status tfsimple -w PipelineReady| jq -M .
```

    {
      "pipelineName": "tfsimple",
      "versions": [
        {
          "pipeline": {
            "name": "tfsimple",
            "uid": "ceevkk6sc0ns73er67qg",
            "version": 1,
            "steps": [
              {
                "name": "tfsimple1"
              }
            ],
            "output": {
              "steps": [
                "tfsimple1.outputs"
              ]
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-12-17T17:20:16.765278466Z",
            "modelsReady": true
          }
        }
      ]
    }



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0,
            0
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



```python
!cat ./pipelines/tfsimple-extended.yaml
!echo "---"
!cat ./pipelines/tfsimple-extended2.yaml
!echo "---"
!cat ./pipelines/tfsimple-combined-step.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended
    spec:
      input:
        externalInputs:
          - tfsimple.outputs
        tensorMap:
          tfsimple.outputs.OUTPUT0: INPUT0
          tfsimple.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-extended2
    spec:
      input:
        externalInputs:
          - tfsimple.outputs
        tensorMap:
          tfsimple.outputs.OUTPUT0: INPUT0
          tfsimple.outputs.OUTPUT1: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-combined-step
    spec:
      input:
        externalInputs:
          - tfsimple-extended.step.tfsimple2.outputs.OUTPUT0
          - tfsimple-extended2.step.tfsimple2.outputs.OUTPUT0
        tensorMap:
          tfsimple-extended.step.tfsimple2.outputs.OUTPUT0: INPUT0
          tfsimple-extended2.step.tfsimple2.outputs.OUTPUT0: INPUT1
      steps:
        - name: tfsimple2
      output:
        steps:
        - tfsimple2



```python
!seldon pipeline load -f ./pipelines/tfsimple-extended.yaml
!seldon pipeline load -f ./pipelines/tfsimple-extended2.yaml
!seldon pipeline load -f ./pipelines/tfsimple-combined-step.yaml
```

    {}
    {}
    {}



```python
!seldon pipeline status tfsimple-extended -w PipelineReady
!seldon pipeline status tfsimple-extended2 -w PipelineReady
!seldon pipeline status tfsimple-combined-step -w PipelineReady
```

    {"pipelineName":"tfsimple-extended", "versions":[{"pipeline":{"name":"tfsimple-extended", "uid":"ceevl7esc0ns73er67r0", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple.outputs"], "tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0", "tfsimple.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:21:33.538468063Z", "modelsReady":true}}]}
    {"pipelineName":"tfsimple-extended2", "versions":[{"pipeline":{"name":"tfsimple-extended2", "uid":"ceevl7esc0ns73er67rg", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple.outputs"], "tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0", "tfsimple.outputs.OUTPUT1":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:21:33.681257781Z", "modelsReady":true}}]}
    {"pipelineName":"tfsimple-combined-step", "versions":[{"pipeline":{"name":"tfsimple-combined-step", "uid":"ceevl7esc0ns73er67s0", "version":1, "steps":[{"name":"tfsimple2"}], "output":{"steps":["tfsimple2.outputs"]}, "kubernetesMeta":{}, "input":{"externalInputs":["tfsimple-extended.step.tfsimple2.outputs.OUTPUT0", "tfsimple-extended2.step.tfsimple2.outputs.OUTPUT0"], "tensorMap":{"tfsimple-extended.step.tfsimple2.outputs.OUTPUT0":"INPUT0", "tfsimple-extended2.step.tfsimple2.outputs.OUTPUT0":"INPUT1"}}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-12-17T17:21:33.911798115Z", "modelsReady":true}}]}



```python
!seldon pipeline infer tfsimple \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' 
```

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
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0,
    				0
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



```python
!seldon pipeline inspect tfsimple
```

    seldon.default.model.tfsimple1.inputs	ceevl847aa7c73d9cftg	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.model.tfsimple1.outputs	ceevl847aa7c73d9cftg	{"modelName":"tfsimple1_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple.inputs	ceevl847aa7c73d9cftg	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]}}]}
    seldon.default.pipeline.tfsimple.outputs	ceevl847aa7c73d9cftg	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}



```python
!seldon pipeline inspect tfsimple-extended
```

    seldon.default.model.tfsimple2.inputs	test-id3	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceepucc7aa7c73d9cfjg	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-extended.inputs	ceevl847aa7c73d9cftg	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended.outputs	ceevl847aa7c73d9cftg	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline inspect tfsimple-extended2
```

    seldon.default.model.tfsimple2.inputs	ceepua47aa7c73d9cfj0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceepucc7aa7c73d9cfjg	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-extended2.inputs	ceevl847aa7c73d9cftg	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended2.outputs	ceevl847aa7c73d9cftg	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}]}



```python
!seldon pipeline inspect tfsimple-combined-step
```

    seldon.default.model.tfsimple2.inputs	ceeq4ls7aa7c73d9cfk0	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	ceepucc7aa7c73d9cfjg	{"modelName":"tfsimple2_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}
    seldon.default.pipeline.tfsimple-combined-step.inputs	ceevl847aa7c73d9cftg	{"inputs":[{"name":"INPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}, {"name":"INPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}}], "rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==", "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.pipeline.tfsimple-combined-step.outputs	ceepua47aa7c73d9cfj0	{"outputs":[{"name":"OUTPUT0", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64]}}, {"name":"OUTPUT1", "datatype":"INT32", "shape":["1", "16"], "contents":{"intContents":[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}}]}



```python
!seldon pipeline unload tfsimple-extended
!seldon pipeline unload tfsimple-extended2
!seldon pipeline unload tfsimple-combined-step
!seldon pipeline unload tfsimple
```

    {}
    {}
    {}
    {}



```python
!seldon model unload tfsimple1
!seldon model unload tfsimple2
```

    {}
    {}



```python

```
