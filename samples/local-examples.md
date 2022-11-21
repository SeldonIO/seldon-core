## Seldon V2 Non Kubernetes Local Examples


### SKLearn Model

We use a simple sklearn iris classification model


```python
!cat ./models/sklearn-iris-gs.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
      memory: 100Ki


Load the model


```python
!seldon model load -f ./models/sklearn-iris-gs.yaml
```

    {}


Wait for the model to be ready


```python
!seldon model status iris -w ModelAvailable | jq -M .
```

    {}


Do a REST inference call


```python
!seldon model infer iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    {
    	"model_name": "iris_1",
    	"model_version": "1",
    	"id": "e08dda16-776e-40e9-b556-273d8bf851ae",
    	"parameters": {
    		"content_type": null,
    		"headers": null
    	},
    	"outputs": [
    		{
    			"name": "predict",
    			"shape": [
    				1
    			],
    			"datatype": "INT64",
    			"parameters": null,
    			"data": [
    				2
    			]
    		}
    	]
    }


Do a gRPC inference call


```python
!seldon model infer iris --inference-mode grpc \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```

    {
      "modelName": "iris_1",
      "modelVersion": "1",
      "outputs": [
        {
          "name": "predict",
          "datatype": "INT64",
          "shape": [
            "1"
          ],
          "contents": {
            "int64Contents": [
              "2"
            ]
          }
        }
      ]
    }


Unload the model


```python
!seldon model unload iris
```

    {}


### Tensorflow Model

We run a simple tensorflow model. Note the requirements section specifying `tensorflow`.


```python
!cat ./models/tfsimple1.yaml
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


Load the model.


```python
!seldon model load -f ./models/tfsimple1.yaml
```

    {}


Wait for the model to be ready.


```python
!seldon model status tfsimple1 -w ModelAvailable | jq -M .
```

    {}


Get model metadata


```python
!seldon model metadata tfsimple1
```

    {
    	"name": "tfsimple1_1",
    	"versions": [
    		"1"
    	],
    	"platform": "tensorflow_graphdef",
    	"inputs": [
    		{
    			"name": "INPUT0",
    			"datatype": "INT32",
    			"shape": [
    				-1,
    				16
    			]
    		},
    		{
    			"name": "INPUT1",
    			"datatype": "INT32",
    			"shape": [
    				-1,
    				16
    			]
    		}
    	],
    	"outputs": [
    		{
    			"name": "OUTPUT0",
    			"datatype": "INT32",
    			"shape": [
    				-1,
    				16
    			]
    		},
    		{
    			"name": "OUTPUT1",
    			"datatype": "INT32",
    			"shape": [
    				-1,
    				16
    			]
    		}
    	]
    }


Do a REST inference call.


```python
!seldon model infer tfsimple1 \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

    {
      "model_name": "tfsimple1_1",
      "model_version": "1",
      "outputs": [
        {
          "name": "OUTPUT0",
          "datatype": "INT32",
          "shape": [
            1,
            16
          ],
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
          ]
        },
        {
          "name": "OUTPUT1",
          "datatype": "INT32",
          "shape": [
            1,
            16
          ],
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
          ]
        }
      ]
    }


Do a gRPC inference call


```python
!seldon model infer tfsimple1 --inference-mode grpc \
    '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

    {
      "modelName": "tfsimple1_1",
      "modelVersion": "1",
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
            ]
          }
        }
      ]
    }


Unload the model


```python
!seldon model unload tfsimple1
```

    {}


### Experiment

We will use two SKlearn Iris classification models to illustrate an experiment.


```python
!cat ./models/sklearn1.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn



```python
!cat ./models/sklearn2.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn


Load both models.


```python
!seldon model load -f ./models/sklearn1.yaml
!seldon model load -f ./models/sklearn2.yaml
```

    {}
    {}


Wait for both models to be ready.


```python
!seldon model status iris | jq -M .
!seldon model status iris2 | jq -M .
```

    {
      "modelName": "iris",
      "versions": [
        {
          "version": 1,
          "serverName": "mlserver",
          "kubernetesMeta": {},
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-11-15T10:45:30.088668539Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-11-15T10:45:30.088668539Z"
          },
          "modelDefn": {
            "meta": {
              "name": "iris",
              "kubernetesMeta": {}
            },
            "modelSpec": {
              "uri": "gs://seldon-models/mlserver/iris",
              "requirements": [
                "sklearn"
              ]
            },
            "deploymentSpec": {
              "replicas": 1
            }
          }
        }
      ]
    }
    {
      "modelName": "iris2",
      "versions": [
        {
          "version": 1,
          "serverName": "mlserver",
          "kubernetesMeta": {},
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-11-15T10:45:30.088816590Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-11-15T10:45:30.088816590Z"
          },
          "modelDefn": {
            "meta": {
              "name": "iris2",
              "kubernetesMeta": {}
            },
            "modelSpec": {
              "uri": "gs://seldon-models/mlserver/iris",
              "requirements": [
                "sklearn"
              ]
            },
            "deploymentSpec": {
              "replicas": 1
            }
          }
        }
      ]
    }


Create an experiment that modifies the iris model to add a second model splitting traffic 50/50 between the two.


```python
!cat ./experiments/ab-default-model.yaml 
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
    spec:
      default: iris
      candidates:
      - name: iris
        weight: 50
      - name: iris2
        weight: 50


Start the experiment.


```python
!seldon experiment start -f ./experiments/ab-default-model.yaml 
```

    {}


Wait for the experiment to be ready.


```python
!seldon experiment status experiment-sample -w | jq -M .
```

    {
      "experimentName": "experiment-sample",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }


Run a set of calls and record which route the traffic took. There should be roughly a 50/50 split.


```python
!seldon model infer iris -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    Success: map[:iris2_1::41 :iris_1::59]


Run one more request


```python
!seldon model infer iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    {
    	"model_name": "iris_1",
    	"model_version": "1",
    	"id": "35a7fce6-db28-4229-ba35-dd2c1db02505",
    	"parameters": {
    		"content_type": null,
    		"headers": null
    	},
    	"outputs": [
    		{
    			"name": "predict",
    			"shape": [
    				1
    			],
    			"datatype": "INT64",
    			"parameters": null,
    			"data": [
    				2
    			]
    		}
    	]
    }


Use sticky session key passed by last infer request to ensure same route is taken each time. 
We will test REST and gRPC.


```python
!seldon model infer iris -s -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    Success: map[:iris_1::50]



```python
!seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' 
```

    Success: map[:iris_1::50]


Stop the experiment


```python
!seldon experiment stop experiment-sample
```

    {}


Show the requests all go to original model now.


```python
!seldon model infer iris -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    Success: map[:iris_1::100]


Unload both models.


```python
!seldon model unload iris
!seldon model unload iris2
```

    {}
    {}



```python

```
