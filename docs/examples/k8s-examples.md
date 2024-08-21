## Seldon V2 Kubernetes Examples

```python
import os
os.environ["NAMESPACE"] = "seldon-mesh"
```

```python
MESH_IP=!kubectl get svc seldon-mesh -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```

```
'172.18.255.2'

```

### Model

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

```bash
kubectl create -f ./models/sklearn-iris-gs.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris condition met

```

```bash
kubectl get model iris -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-06-30T10:01:52Z",
      "message": "ModelAvailable",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-06-30T10:01:52Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "7fd401e1-3dce-46f5-9668-902aea652b89",
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

```bash
seldon model infer iris --inference-mode grpc --inference-host ${MESH_IP}:80 \
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

```bash
kubectl get server mlserver -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-06-30T09:59:12Z",
      "status": "True",
      "type": "Ready"
    },
    {
      "lastTransitionTime": "2023-06-30T09:59:12Z",
      "reason": "StatefulSet replicas matches desired replicas",
      "status": "True",
      "type": "StatefulSetReady"
    }
  ],
  "loadedModels": 1,
  "replicas": 1,
  "selector": "seldon-server-name=mlserver"
}

```

```bash
kubectl delete -f ./models/sklearn-iris-gs.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "iris" deleted

```

### Experiment

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

```bash
kubectl create -f ./models/sklearn1.yaml -n ${NAMESPACE}
kubectl create -f ./models/sklearn2.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris created
model.mlops.seldon.io/iris2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris condition met
model.mlops.seldon.io/iris2 condition met

```

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

```bash
kubectl create -f ./experiments/ab-default-model.yaml -n ${NAMESPACE}
```

```
experiment.mlops.seldon.io/experiment-sample created

```

```bash
kubectl wait --for condition=ready --timeout=300s experiment --all -n ${NAMESPACE}
```

```
experiment.mlops.seldon.io/experiment-sample condition met

```

```bash
seldon model infer --inference-host ${MESH_IP}:80 -i 50 iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::29 :iris_1::21]

```

```bash
kubectl delete -f ./experiments/ab-default-model.yaml -n ${NAMESPACE}
kubectl delete -f ./models/sklearn1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/sklearn2.yaml -n ${NAMESPACE}
```

```
experiment.mlops.seldon.io "experiment-sample" deleted
model.mlops.seldon.io "iris" deleted
model.mlops.seldon.io "iris2" deleted

```

### Pipeline - model chain

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
kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met

```

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
kubectl create -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/tfsimples created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/tfsimples condition met

```

```bash
seldon pipeline infer tfsimples --inference-mode grpc --inference-host ${MESH_IP}:80 \
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
kubectl delete -f ./pipelines/tfsimples.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io "tfsimples" deleted

```

```bash
kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted

```

### Pipeline - model join

```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
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
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple2
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
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
kubectl create -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl create -f ./models/tfsimple2.yaml -n ${NAMESPACE}
kubectl create -f ./models/tfsimple3.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/tfsimple1 created
model.mlops.seldon.io/tfsimple2 created
model.mlops.seldon.io/tfsimple3 created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/tfsimple1 condition met
model.mlops.seldon.io/tfsimple2 condition met
model.mlops.seldon.io/tfsimple3 condition met

```

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
kubectl create -f ./pipelines/tfsimples-join.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/join created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/join condition met

```

```bash
seldon pipeline infer join --inference-mode grpc --inference-host ${MESH_IP}:80 \
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
kubectl delete -f ./pipelines/tfsimples-join.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io "join" deleted

```

```bash
kubectl delete -f ./models/tfsimple1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/tfsimple2.yaml -n ${NAMESPACE}
kubectl delete -f ./models/tfsimple3.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "tfsimple1" deleted
model.mlops.seldon.io "tfsimple2" deleted
model.mlops.seldon.io "tfsimple3" deleted

```

## Explainer

```bash
cat ./models/income.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/income/classifier"
  requirements:
  - sklearn

```

```bash
kubectl create -f ./models/income.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income condition met

```

```bash
kubectl get model income -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-06-30T10:02:53Z",
      "message": "ModelAvailable",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-06-30T10:02:53Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer income --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income_1",
	"model_version": "1",
	"id": "f52acfeb-0f22-429f-8c7a-785ef17cd470",
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
				0
			]
		}
	]
}

```

```bash
cat ./models/income-explainer.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-explainer
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/income/explainer"
  explainer:
    type: anchor_tabular
    modelRef: income

```

```bash
kubectl create -f ./models/income-explainer.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income-explainer created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/income condition met
model.mlops.seldon.io/income-explainer condition met

```

```bash
kubectl get model income-explainer -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

```json
{
  "conditions": [
    {
      "lastTransitionTime": "2023-06-30T10:03:07Z",
      "message": "ModelAvailable",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-06-30T10:03:07Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer income-explainer --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income-explainer_1",
	"model_version": "1",
	"id": "3028a904-9bb3-42d7-bdb7-6e6993323ed7",
	"parameters": {},
	"outputs": [
		{
			"name": "explanation",
			"shape": [
				1,
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "str"
			},
			"data": [
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.9.1\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\"], \"precision\": 0.9705882352941176, \"coverage\": 0.0699, \"raw\": {\"feature\": [3, 5], \"mean\": [0.8094218415417559, 0.9705882352941176], \"precision\": [0.8094218415417559, 0.9705882352941176], \"coverage\": [0.3036, 0.0699], \"examples\": [{\"covered_true\": [[23, 4, 1, 1, 5, 1, 4, 0, 0, 0, 40, 9], [44, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [60, 2, 5, 1, 5, 1, 4, 0, 0, 0, 25, 9], [52, 4, 1, 1, 2, 0, 4, 1, 0, 0, 50, 9], [66, 6, 1, 1, 8, 0, 4, 1, 0, 0, 8, 9], [52, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [27, 4, 1, 1, 1, 1, 4, 1, 0, 0, 35, 9], [48, 4, 1, 1, 6, 0, 4, 1, 0, 0, 45, 9], [45, 6, 1, 1, 5, 0, 4, 1, 0, 0, 40, 9], [40, 2, 1, 1, 5, 4, 4, 0, 0, 0, 45, 9]], \"covered_false\": [[42, 6, 5, 1, 6, 0, 4, 1, 99999, 0, 80, 9], [29, 4, 1, 1, 8, 1, 4, 1, 0, 0, 50, 9], [49, 4, 1, 1, 8, 0, 4, 1, 0, 0, 50, 9], [34, 4, 5, 1, 8, 0, 4, 1, 0, 0, 40, 9], [38, 2, 1, 1, 5, 5, 4, 0, 7688, 0, 40, 9], [45, 7, 5, 1, 5, 0, 4, 1, 0, 0, 45, 9], [43, 4, 2, 1, 5, 0, 4, 1, 99999, 0, 55, 9], [47, 4, 5, 1, 6, 1, 4, 1, 27828, 0, 60, 9], [42, 6, 1, 1, 2, 0, 4, 1, 15024, 0, 60, 9], [56, 4, 1, 1, 6, 0, 2, 1, 7688, 0, 45, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[23, 4, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [50, 2, 5, 1, 8, 3, 2, 1, 0, 0, 45, 9], [24, 4, 1, 1, 7, 3, 4, 0, 0, 0, 40, 3], [62, 4, 5, 1, 5, 3, 4, 1, 0, 0, 40, 9], [22, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [44, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [46, 4, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [44, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [25, 4, 5, 1, 5, 3, 4, 1, 0, 0, 35, 9], [32, 2, 5, 1, 5, 3, 4, 1, 0, 0, 50, 9]], \"covered_false\": [[57, 5, 5, 1, 6, 3, 4, 1, 99999, 0, 40, 9], [44, 4, 1, 1, 8, 3, 4, 1, 7688, 0, 60, 9], [43, 2, 5, 1, 4, 3, 2, 0, 8614, 0, 47, 9], [56, 5, 2, 1, 5, 3, 4, 1, 99999, 0, 70, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
			]
		}
	]
}

```

```bash
kubectl delete -f ./models/income.yaml -n ${NAMESPACE}
kubectl delete -f ./models/income-explainer.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "income" deleted
model.mlops.seldon.io "income-explainer" deleted

```
