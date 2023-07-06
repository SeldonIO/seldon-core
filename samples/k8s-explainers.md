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
'172.19.255.1'

```

## Explain Model

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
      "lastTransitionTime": "2023-03-10T10:54:55Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-03-10T10:54:55Z",
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
	"id": "3fa3a933-37dd-4e42-b5fd-f5bbc56f8bf4",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
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
      "lastTransitionTime": "2023-03-10T10:55:07Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-03-10T10:55:07Z",
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
	"id": "55af3c9b-defd-4a52-8047-cc532461e8a4",
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
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.9.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9974489795918368, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.7767175572519084, 0.9107142857142857, 0.9974489795918368], \"precision\": [0.7767175572519084, 0.9107142857142857, 0.9974489795918368], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[25, 2, 1, 1, 1, 0, 4, 1, 0, 0, 40, 9], [23, 4, 1, 1, 1, 3, 2, 1, 0, 0, 24, 9], [25, 4, 1, 1, 2, 0, 4, 1, 0, 0, 60, 9], [38, 4, 1, 1, 5, 0, 4, 1, 0, 0, 40, 9], [39, 4, 5, 1, 5, 5, 4, 0, 0, 1902, 18, 9], [24, 4, 5, 1, 5, 5, 4, 0, 0, 0, 35, 9], [23, 4, 5, 1, 5, 0, 1, 1, 0, 1887, 50, 1], [39, 5, 1, 1, 6, 1, 4, 1, 0, 0, 40, 9], [34, 4, 5, 1, 5, 0, 4, 1, 0, 0, 55, 9], [46, 5, 1, 1, 8, 1, 4, 0, 0, 0, 50, 9]], \"covered_false\": [[27, 4, 1, 1, 6, 1, 4, 1, 0, 0, 80, 9], [47, 4, 1, 1, 4, 0, 4, 1, 15024, 0, 40, 9], [38, 4, 5, 1, 8, 0, 4, 1, 99999, 0, 70, 9], [43, 4, 5, 1, 8, 0, 4, 1, 0, 0, 50, 9], [47, 4, 1, 1, 8, 0, 4, 1, 3103, 0, 60, 9], [46, 4, 1, 1, 8, 0, 4, 1, 15024, 0, 45, 9], [36, 4, 5, 1, 8, 0, 4, 1, 15024, 0, 50, 9], [45, 4, 5, 1, 8, 0, 4, 1, 0, 0, 50, 9], [37, 4, 5, 1, 8, 0, 4, 1, 0, 0, 50, 9], [43, 1, 1, 1, 8, 1, 4, 1, 0, 0, 40, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[36, 4, 1, 1, 7, 3, 4, 0, 0, 0, 40, 9], [45, 2, 5, 1, 5, 3, 4, 0, 0, 0, 45, 9], [34, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [42, 2, 5, 1, 5, 3, 4, 0, 0, 0, 45, 9], [44, 4, 1, 1, 4, 3, 4, 1, 5178, 0, 40, 9], [35, 4, 5, 1, 1, 3, 4, 1, 0, 0, 40, 9], [25, 0, 5, 1, 0, 3, 4, 0, 0, 0, 35, 1], [27, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [26, 2, 5, 1, 5, 3, 4, 1, 0, 0, 50, 9], [36, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[25, 4, 1, 1, 2, 3, 4, 1, 27828, 0, 50, 9], [39, 4, 1, 1, 8, 3, 4, 1, 15024, 0, 50, 9], [44, 4, 1, 1, 5, 3, 4, 1, 15024, 0, 50, 3], [36, 7, 1, 1, 4, 3, 4, 1, 7298, 0, 55, 9], [38, 6, 1, 1, 6, 3, 4, 1, 7298, 0, 50, 9], [50, 4, 5, 1, 5, 3, 4, 1, 15024, 0, 45, 9], [39, 4, 1, 1, 6, 3, 4, 1, 15024, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[28, 4, 5, 1, 6, 3, 4, 0, 0, 0, 80, 9], [29, 2, 1, 1, 5, 3, 4, 1, 0, 0, 60, 9], [57, 2, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [41, 4, 1, 1, 2, 3, 2, 1, 0, 0, 40, 0], [32, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [40, 4, 1, 1, 2, 3, 3, 1, 0, 0, 45, 0], [34, 4, 1, 1, 4, 3, 2, 1, 0, 0, 30, 9], [50, 0, 5, 1, 0, 3, 4, 1, 0, 0, 40, 9], [53, 5, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [23, 4, 1, 1, 8, 3, 4, 0, 0, 0, 15, 9]], \"covered_false\": [[23, 6, 1, 1, 1, 3, 4, 1, 0, 2231, 40, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
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

## Explain Pipeline

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
      "lastTransitionTime": "2023-03-10T10:55:15Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-03-10T10:55:15Z",
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
	"id": "a7d69896-1396-4c15-92bb-bd586cb3572a",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"data": [
				0
			]
		}
	]
}

```

```bash
cat ./pipelines/income-v1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: income-prod
spec:
  steps:
    - name: income
  output:
    steps:
    - income

```

```bash
kubectl create -f ./pipelines/income-v1.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/income-prod created

```

```bash
kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io/income-prod condition met

```

```bash
seldon pipeline infer income-prod --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "",
	"outputs": [
		{
			"data": [
				0
			],
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64"
		}
	]
}

```

```bash
cat ./models/income-explainer-pipeline.yaml
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
    pipelineRef: income-prod

```

```bash
kubectl create -f ./models/income-explainer-pipeline.yaml -n ${NAMESPACE}
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
      "lastTransitionTime": "2023-03-10T10:55:34Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2023-03-10T10:55:34Z",
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
	"id": "aff2498f-c3be-4056-b718-eaa2f714d75b",
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
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.9.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9937629937629938, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.7935606060606061, 0.9120879120879121, 0.9937629937629938], \"precision\": [0.7935606060606061, 0.9120879120879121, 0.9937629937629938], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[27, 4, 1, 1, 1, 1, 4, 0, 0, 0, 65, 9], [35, 4, 1, 1, 6, 0, 4, 1, 0, 0, 60, 9], [40, 4, 1, 1, 2, 0, 4, 1, 0, 0, 40, 9], [63, 0, 2, 1, 0, 0, 4, 1, 0, 0, 12, 9], [40, 0, 1, 1, 0, 1, 4, 0, 0, 0, 50, 9], [24, 4, 1, 1, 2, 3, 2, 0, 0, 0, 40, 9], [36, 4, 1, 1, 5, 4, 4, 0, 0, 0, 44, 9], [24, 4, 1, 1, 6, 3, 2, 1, 0, 0, 40, 9], [48, 4, 1, 1, 7, 4, 1, 0, 0, 0, 40, 7], [39, 4, 1, 1, 2, 0, 1, 1, 0, 0, 40, 7]], \"covered_false\": [[32, 4, 5, 1, 8, 0, 2, 1, 0, 1977, 40, 9], [36, 4, 1, 1, 8, 1, 4, 1, 8614, 0, 40, 9], [47, 2, 5, 1, 8, 0, 2, 1, 0, 1977, 50, 9], [48, 4, 1, 1, 8, 0, 4, 1, 0, 0, 45, 9], [49, 4, 5, 1, 5, 1, 4, 1, 0, 0, 40, 9], [48, 7, 2, 1, 5, 0, 4, 1, 0, 0, 65, 9], [38, 4, 5, 1, 5, 0, 4, 1, 99999, 0, 65, 9], [49, 2, 5, 1, 8, 0, 2, 1, 0, 0, 47, 9], [62, 5, 5, 1, 6, 0, 4, 1, 0, 0, 45, 9], [46, 7, 2, 1, 5, 0, 4, 1, 7688, 0, 45, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[43, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [22, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [60, 2, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [57, 4, 5, 1, 5, 3, 4, 1, 0, 1977, 45, 9], [65, 4, 5, 1, 6, 3, 4, 1, 0, 0, 28, 9], [50, 2, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [48, 4, 1, 1, 8, 3, 1, 1, 0, 0, 40, 1], [46, 7, 1, 1, 8, 3, 4, 1, 0, 0, 45, 9], [48, 2, 5, 1, 5, 3, 4, 0, 0, 0, 80, 9], [22, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[30, 4, 1, 1, 5, 3, 4, 1, 99999, 0, 35, 9], [52, 4, 5, 1, 5, 3, 4, 1, 15020, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[38, 4, 1, 1, 2, 3, 2, 1, 0, 0, 40, 5], [35, 4, 5, 1, 5, 3, 4, 1, 0, 1887, 40, 9], [45, 4, 2, 1, 5, 3, 2, 0, 0, 3004, 35, 9], [78, 5, 1, 1, 8, 3, 4, 1, 0, 2392, 40, 9], [33, 4, 1, 1, 8, 3, 4, 1, 0, 0, 55, 9], [29, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [51, 2, 1, 1, 5, 3, 4, 1, 0, 0, 26, 9], [48, 4, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [45, 2, 5, 1, 5, 3, 4, 0, 0, 0, 30, 9], [29, 4, 5, 1, 8, 3, 1, 1, 0, 0, 40, 6]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
			]
		}
	]
}

```

```bash
kubectl delete -f ./pipelines/income-v1.yaml -n ${NAMESPACE}
kubectl delete -f ./models/income-explainer-pipeline.yaml -n ${NAMESPACE}
kubectl delete -f ./models/income.yaml -n ${NAMESPACE}
```

```
pipeline.mlops.seldon.io "income-prod" deleted
model.mlops.seldon.io "income-explainer" deleted
model.mlops.seldon.io "income" deleted

```

```python

```
