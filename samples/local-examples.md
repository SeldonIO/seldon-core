## Seldon V2 Non Kubernetes Local Examples


 * Build if needed and place `seldon` binary in your path
   * run `make build-seldon` from operator folder and add bin folder to `PATH`
 * Run Seldon V2 `make deploy-local` from top level folder

### SKLearn Model

We use a simple sklearn iris classification model


```bash
cat ./models/sklearn-iris-gs.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````
Load the model


```bash
seldon model load -f ./models/sklearn-iris-gs.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````
Wait for the model to be ready


```bash
seldon model status --model-name iris -w ModelAvailable | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "iris",
      "versions": [
        {
          "version": 1,
          "serverName": "mlserver",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-05-07T08:24:44.069876241Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-05-07T08:24:44.069876241Z"
          },
          "modelDefn": {
            "meta": {
              "name": "iris",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/mlserver/iris",
              "requirements": [
                "sklearn"
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
Do a REST inference call


```bash
seldon model infer --model-name iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "iris_1",
    	"model_version": "1",
    	"id": "cb099f86-9426-4a76-9d33-7faed0ac72b3",
    	"parameters": null,
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
```
````
Do a gRPC inference call


```bash
seldon model infer --model-name iris --inference-mode grpc \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

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
```
````
Unload the model


```bash
seldon model unload --model-name iris
```
````{collapse} Expand to see output
```json

    {}
```
````
### Tensorflow Model


```bash
cat ./models/tfsimple1.yaml
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
```
````
Load the model.


```bash
seldon model load -f ./models/tfsimple1.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````
Wait for the model to be ready.


```bash
seldon model status --model-name tfsimple1 -w ModelAvailable | jq -M .
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
              "lastChangeTimestamp": "2022-05-07T08:30:48.321215481Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-05-07T08:30:48.321215481Z"
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
```
````
Get model metadata


```bash
seldon model metadata -m tfsimple1
```
````{collapse} Expand to see output
```json

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
```
````
Do a REST inference call.


```bash
seldon model infer -m tfsimple1 \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

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
```
````
Do a gRPC inference call


```bash
seldon model infer -m tfsimple1 --inference-mode grpc \
    '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
````{collapse} Expand to see output
```json

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
      ],
      "rawOutputContents": [
        "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
        "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
      ]
    }
```
````
Unload the model


```bash
seldon model unload --model-name tfsimple1
```
````{collapse} Expand to see output
```json

    {}
```
````
### Experiment

We will use two SKlearn Iris classification models to illustrate experiments.


```bash
cat ./experiments/sklearn1.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
cat ./experiments/sklearn2.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````
Load both models.


```bash
seldon model load -f ./experiments/sklearn1.yaml
seldon model load -f ./experiments/sklearn2.yaml
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
Wait for both models to be ready.


```bash
seldon model status --model-name iris | jq -M .
seldon model status --model-name iris2 | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "modelName": "iris",
      "versions": [
        {
          "version": 1,
          "serverName": "mlserver",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-05-07T08:31:54.494645822Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-05-07T08:31:54.494645822Z"
          },
          "modelDefn": {
            "meta": {
              "name": "iris",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/mlserver/iris",
              "requirements": [
                "sklearn"
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
      "modelName": "iris2",
      "versions": [
        {
          "version": 1,
          "serverName": "mlserver",
          "kubernetesMeta": {
            "namespace": "seldon-mesh"
          },
          "modelReplicaState": {
            "0": {
              "state": "Available",
              "lastChangeTimestamp": "2022-05-07T08:31:54.517961482Z"
            }
          },
          "state": {
            "state": "ModelAvailable",
            "availableReplicas": 1,
            "lastChangeTimestamp": "2022-05-07T08:31:54.517961482Z"
          },
          "modelDefn": {
            "meta": {
              "name": "iris2",
              "kubernetesMeta": {
                "namespace": "seldon-mesh"
              }
            },
            "modelSpec": {
              "uri": "gs://seldon-models/mlserver/iris",
              "requirements": [
                "sklearn"
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
Create an experiment that modifies the iris model to add a second model splitting traffic 50/50 between the two.


```bash
cat ./experiments/ab-default-model.yaml 
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
      namespace: seldon-mesh
    spec:
      defaultModel: iris
      candidates:
      - modelName: iris
        weight: 50
      - modelName: iris2
        weight: 50
```
````
Start the experiment.


```bash
seldon experiment start -f ./experiments/ab-default-model.yaml 
```
````{collapse} Expand to see output
```json

    {}
```
````
Wait for the experiment to be ready.


```bash
seldon experiment status -e experiment-sample -w | jq -M .
```
````{collapse} Expand to see output
```json

    {
      "experimentName": "experiment-sample",
      "active": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {
        "namespace": "seldon-mesh"
      }
    }
```
````
Run a set of calls and record which route the traffic took. There should be roughly a 50/50 split.


```bash
seldon model infer --model-name iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output
```json

    map[iris2_1:24 iris_1:26]
```
````
Stop the experiment


```bash
seldon experiment stop -e experiment-sample
```
````{collapse} Expand to see output
```json

    {}
```
````
Unload both models.


```bash
seldon model unload --model-name iris
seldon model unload --model-name iris2
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````