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
            "uid": "cebnb8ev219s73a6ojk0",
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
            "lastChangeTimestamp": "2022-12-12T18:40:33.304154775Z",
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
            "uid": "cebnb9mv219s73a6ojkg",
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
            "lastChangeTimestamp": "2022-12-12T18:40:39.004261615Z",
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

    seldon.default.model.tfsimple1.inputs	cebnba9q03gs739pd0eg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    seldon.default.model.tfsimple1.outputs	cebnba9q03gs739pd0eg	{"modelName":"tfsimple1_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple.inputs	cebnba9q03gs739pd0eg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    seldon.default.pipeline.tfsimple.outputs	cebnba9q03gs739pd0eg	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}



```python
!seldon pipeline inspect tfsimple-extended
```

    seldon.default.model.tfsimple2.inputs	cebnba9q03gs739pd0eg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	cebnba9q03gs739pd0eg	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}
    seldon.default.pipeline.tfsimple-extended.inputs	cebnba9q03gs739pd0eg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended.outputs	cebnba9q03gs739pd0eg	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}



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
            "uid": "cebnacuv219s73a6oji0",
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
            "lastChangeTimestamp": "2022-12-12T18:38:43.317204867Z",
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

    {"pipelineName":"tfsimple-extended","versions":[{"pipeline":{"name":"tfsimple-extended","uid":"cebnaemv219s73a6ojig","version":1,"steps":[{"name":"tfsimple2"}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{},"input":{"externalInputs":["tfsimple.outputs"],"tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0","tfsimple.outputs.OUTPUT1":"INPUT1"}}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-12-12T18:38:50.324389094Z","modelsReady":true}}]}
    {"pipelineName":"tfsimple-extended2","versions":[{"pipeline":{"name":"tfsimple-extended2","uid":"cebnaemv219s73a6ojj0","version":1,"steps":[{"name":"tfsimple2"}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{},"input":{"externalInputs":["tfsimple.outputs"],"tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0","tfsimple.outputs.OUTPUT1":"INPUT1"}}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-12-12T18:38:50.467839593Z","modelsReady":true}}]}
    {"pipelineName":"tfsimple-combined","versions":[{"pipeline":{"name":"tfsimple-combined","uid":"cebnaemv219s73a6ojjg","version":2,"steps":[{"name":"tfsimple2"}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{},"input":{"externalInputs":["tfsimple-extended.outputs.OUTPUT0","tfsimple-extended2.outputs.OUTPUT1"],"tensorMap":{"tfsimple-extended.outputs.OUTPUT0":"INPUT0","tfsimple-extended2.outputs.OUTPUT1":"INPUT1"}}},"state":{"pipelineVersion":2,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-12-12T18:38:50.718933075Z","modelsReady":true}}]}



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

    seldon.default.model.tfsimple1.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    seldon.default.model.tfsimple1.outputs	cebnaf9q03gs739pd0dg	{"modelName":"tfsimple1_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    seldon.default.pipeline.tfsimple.outputs	cebnaf9q03gs739pd0dg	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}



```python
!seldon pipeline inspect tfsimple-extended
```

    seldon.default.model.tfsimple2.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	cebnaf9q03gs739pd0dg	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple-extended.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended.outputs	cebnaf9q03gs739pd0dg	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}



```python
!seldon pipeline inspect tfsimple-extended2
```

    seldon.default.model.tfsimple2.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	cebnaf9q03gs739pd0dg	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple-extended2.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended2.outputs	cebnaf9q03gs739pd0dg	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}



```python
!seldon pipeline inspect tfsimple-combined
```

    seldon.default.model.tfsimple2.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	cebnaf9q03gs739pd0dg	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple-combined.inputs	cebnaf9q03gs739pd0dg	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.pipeline.tfsimple-combined.outputs	cebnaf9q03gs739pd0dg	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}



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
            "uid": "cebnbguv219s73a6ojl0",
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
            "lastChangeTimestamp": "2022-12-12T18:41:07.131448326Z",
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

    {"pipelineName":"tfsimple-extended","versions":[{"pipeline":{"name":"tfsimple-extended","uid":"cebnbimv219s73a6ojlg","version":1,"steps":[{"name":"tfsimple2"}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{},"input":{"externalInputs":["tfsimple.outputs"],"tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0","tfsimple.outputs.OUTPUT1":"INPUT1"}}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-12-12T18:41:14.561905446Z","modelsReady":true}}]}
    {"pipelineName":"tfsimple-extended2","versions":[{"pipeline":{"name":"tfsimple-extended2","uid":"cebnbimv219s73a6ojm0","version":1,"steps":[{"name":"tfsimple2"}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{},"input":{"externalInputs":["tfsimple.outputs"],"tensorMap":{"tfsimple.outputs.OUTPUT0":"INPUT0","tfsimple.outputs.OUTPUT1":"INPUT1"}}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-12-12T18:41:14.693285174Z","modelsReady":true}}]}
    {"pipelineName":"tfsimple-combined-trigger","versions":[{"pipeline":{"name":"tfsimple-combined-trigger","uid":"cebnbimv219s73a6ojmg","version":1,"steps":[{"name":"tfsimple2"}],"output":{"steps":["tfsimple2.outputs"]},"kubernetesMeta":{},"input":{"externalInputs":["tfsimple-extended.outputs"],"externalTriggers":["tfsimple-extended2.outputs"],"tensorMap":{"tfsimple-extended.outputs.OUTPUT0":"INPUT0","tfsimple-extended.outputs.OUTPUT1":"INPUT1"}}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-12-12T18:41:14.945627389Z","modelsReady":true}}]}



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

    seldon.default.model.tfsimple1.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    seldon.default.model.tfsimple1.outputs	test-id3	{"modelName":"tfsimple1_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]}}]}
    seldon.default.pipeline.tfsimple.outputs	test-id3	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}



```python
!seldon pipeline inspect tfsimple-extended
```

    seldon.default.model.tfsimple2.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	test-id1	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple-extended.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended.outputs	test-id3	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}



```python
!seldon pipeline inspect tfsimple-extended2
```

    seldon.default.model.tfsimple2.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	test-id1	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple-extended2.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="]}
    seldon.default.pipeline.tfsimple-extended2.outputs	test-id3	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}]}



```python
!seldon pipeline inspect tfsimple-combined-trigger
```

    seldon.default.model.tfsimple2.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.model.tfsimple2.outputs	cebki39q03gs739pd0ag	{"modelName":"tfsimple2_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}
    seldon.default.pipeline.tfsimple-combined-trigger.inputs	test-id3	{"inputs":[{"name":"INPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"INPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}}],"rawInputContents":["AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==","AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA=="]}
    seldon.default.pipeline.tfsimple-combined-trigger.outputs	cebjfl1q03gs739pd09g	{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[4,8,12,16,20,24,28,32,36,40,44,48,52,56,60,64]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}



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



```python

```
