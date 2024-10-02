# Seldon V2 Non Kubernetes Local Examples

### SKLearn Model

We use a simple sklearn iris classification model

```bash
cat ./models/sklearn-iris-gs.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn"
  requirements:
  - sklearn
  memory: 100Ki

```

Load the model

```bash
seldon model load -f ./models/sklearn-iris-gs.yaml
```

```json
{}

```

Wait for the model to be ready

```bash
seldon model status iris -w ModelAvailable | jq -M .
```

```json
{}

```

Do a REST inference call

```bash
seldon model infer iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "983bd95f-4b4d-4ff1-95b2-df9d6d089164",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				2
			]
		}
	]
}

```

Do a gRPC inference call

```bash
seldon model infer iris --inference-mode grpc \
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' | jq -M .
```

```json
{
  "modelName": "iris_1",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "predict",
      "datatype": "INT64",
      "shape": [
        "1",
        "1"
      ],
      "parameters": {
        "content_type": {
          "stringParam": "np"
        }
      },
      "contents": {
        "int64Contents": [
          "2"
        ]
      }
    }
  ]
}

```

Unload the model

```bash
seldon model unload iris
```

### Tensorflow Model

We run a simple tensorflow model. Note the requirements section specifying `tensorflow`.

```bash
cat ./models/tfsimple1.yaml
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

```

Load the model.

```bash
seldon model load -f ./models/tfsimple1.yaml
```

```json
{}

```

Wait for the model to be ready.

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
```

```json
{}

```

Get model metadata

```bash
seldon model metadata tfsimple1
```

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

Do a REST inference call.

```bash
seldon model infer tfsimple1 \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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

Do a gRPC inference call

```bash
seldon model infer tfsimple1 --inference-mode grpc \
    '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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
  ]
}

```

Unload the model

```bash
seldon model unload tfsimple1
```

### Experiment

We will use two SKlearn Iris classification models to illustrate an experiment.

```bash
cat ./models/sklearn1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn

```

```bash
cat ./models/sklearn2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris2
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn

```

Load both models.

```bash
seldon model load -f ./models/sklearn1.yaml
seldon model load -f ./models/sklearn2.yaml
```

```json
{}
{}

```

Wait for both models to be ready.

```bash
seldon model status iris | jq -M .
seldon model status iris2 | jq -M .
```

```json
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
          "lastChangeTimestamp": "2023-06-29T14:01:41.362720538Z"
        }
      },
      "state": {
        "state": "ModelAvailable",
        "availableReplicas": 1,
        "lastChangeTimestamp": "2023-06-29T14:01:41.362720538Z"
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
          "lastChangeTimestamp": "2023-06-29T14:01:41.362845079Z"
        }
      },
      "state": {
        "state": "ModelAvailable",
        "availableReplicas": 1,
        "lastChangeTimestamp": "2023-06-29T14:01:41.362845079Z"
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

```

Create an experiment that modifies the iris model to add a second model splitting traffic 50/50 between the two.

```bash
cat ./experiments/ab-default-model.yaml
```

```yaml
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

```

Start the experiment.

```bash
seldon experiment start -f ./experiments/ab-default-model.yaml
```

Wait for the experiment to be ready.

```bash
seldon experiment status experiment-sample -w | jq -M .
```

```json
{
  "experimentName": "experiment-sample",
  "active": true,
  "candidatesReady": true,
  "mirrorReady": true,
  "statusDescription": "experiment active",
  "kubernetesMeta": {}
}

```

Run a set of calls and record which route the traffic took. There should be roughly a 50/50 split.

```bash
seldon model infer iris -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::57 :iris_1::43]

```

Run one more request

```bash
seldon model infer iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "fa425bdf-737c-41fe-894d-58868f70fe5d",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				2
			]
		}
	]
}

```

Use sticky session key passed by last infer request to ensure same route is taken each time.
We will test REST and gRPC.

```bash
seldon model infer iris -s -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
```

```
Success: map[:iris_1::50]

```

Stop the experiment

```bash
seldon experiment stop experiment-sample
```

Show the requests all go to original model now.

```bash
seldon model infer iris -i 100 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::100]

```

Unload both models.

```bash
seldon model unload iris
seldon model unload iris2
```
