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
  storageUri: "gs://seldon-models/scv2/examples/income/classifier"
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
      "lastTransitionTime": "2022-11-24T11:54:17Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2022-11-24T11:54:17Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer income --inference-host ${MESH_IP}:80 \
```yaml
'{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```
```

```json
{
	"model_name": "income_1",
	"model_version": "1",
	"id": "059ef7a5-b353-4487-a5f2-9c4bccbf7672",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": null,
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
  storageUri: "gs://seldon-models/scv2/examples/income/explainer"
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
      "lastTransitionTime": "2022-11-24T11:54:27Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2022-11-24T11:54:27Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer income-explainer --inference-host ${MESH_IP}:80 \
```yaml
'{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```
```

```json
{
	"model_name": "income-explainer_1",
	"model_version": "1",
	"id": "a43a931c-5b97-4e71-b9ed-35cb728af4ae",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "explanation",
			"shape": [
				1,
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "str",
				"headers": null
			},
			"data": [
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9970238095238095, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.7781818181818182, 0.9153439153439153, 0.9970238095238095], \"precision\": [0.7781818181818182, 0.9153439153439153, 0.9970238095238095], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[54, 4, 1, 1, 6, 4, 4, 0, 0, 0, 40, 9], [49, 2, 5, 1, 8, 0, 4, 1, 0, 0, 40, 9], [32, 0, 1, 1, 0, 5, 4, 0, 0, 0, 20, 9], [26, 4, 1, 1, 5, 1, 4, 0, 0, 0, 40, 9], [45, 4, 1, 1, 5, 1, 4, 1, 0, 0, 46, 9], [31, 4, 1, 1, 8, 1, 4, 1, 0, 0, 40, 9], [23, 4, 1, 1, 1, 1, 1, 0, 0, 0, 20, 9], [47, 6, 1, 1, 2, 1, 4, 0, 0, 0, 40, 9], [43, 6, 1, 1, 2, 0, 4, 1, 0, 0, 70, 9], [29, 7, 5, 1, 5, 5, 4, 0, 0, 0, 45, 9]], \"covered_false\": [[35, 4, 1, 1, 7, 0, 0, 1, 7688, 0, 20, 9], [35, 4, 1, 1, 8, 0, 4, 1, 7688, 0, 50, 9], [51, 4, 5, 1, 8, 0, 4, 1, 7298, 0, 50, 9], [46, 2, 1, 1, 5, 3, 4, 1, 4787, 0, 45, 9], [35, 4, 1, 1, 6, 0, 4, 1, 15024, 0, 50, 9], [29, 4, 1, 1, 8, 1, 4, 1, 0, 0, 50, 9], [31, 4, 5, 1, 8, 1, 4, 1, 0, 1564, 40, 9], [57, 7, 2, 1, 5, 0, 4, 1, 0, 0, 40, 9], [46, 4, 5, 1, 5, 0, 4, 1, 0, 0, 40, 9], [48, 5, 1, 1, 8, 0, 4, 1, 0, 0, 60, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[43, 4, 1, 1, 5, 3, 4, 1, 0, 1902, 40, 9], [34, 5, 1, 1, 5, 3, 4, 0, 0, 0, 62, 9], [23, 4, 1, 1, 1, 3, 1, 0, 0, 0, 20, 9], [49, 4, 5, 1, 5, 3, 4, 1, 0, 0, 40, 9], [31, 2, 1, 1, 6, 3, 4, 0, 0, 0, 40, 9], [48, 4, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [37, 4, 1, 1, 8, 3, 4, 1, 0, 0, 45, 9], [38, 4, 1, 1, 5, 3, 4, 1, 0, 1887, 40, 9], [41, 6, 1, 1, 2, 3, 4, 1, 0, 0, 35, 9], [47, 4, 1, 1, 5, 3, 4, 1, 0, 0, 50, 9]], \"covered_false\": [[50, 4, 5, 1, 8, 3, 4, 1, 15024, 0, 65, 9], [51, 4, 1, 1, 6, 3, 4, 1, 7688, 0, 50, 9], [33, 4, 5, 1, 5, 3, 4, 1, 15024, 0, 44, 9], [42, 4, 1, 1, 5, 3, 4, 1, 7688, 0, 40, 9], [44, 4, 1, 1, 8, 3, 4, 1, 15024, 0, 45, 9], [34, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [51, 4, 1, 1, 5, 3, 0, 1, 15024, 0, 40, 9], [43, 4, 1, 1, 8, 3, 4, 1, 15024, 0, 50, 9], [53, 7, 2, 1, 5, 3, 4, 1, 7688, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[62, 7, 5, 1, 5, 3, 4, 1, 0, 0, 60, 9], [38, 6, 1, 1, 5, 3, 1, 1, 0, 0, 60, 9], [32, 1, 1, 1, 8, 3, 4, 0, 0, 0, 60, 9], [22, 4, 1, 1, 1, 3, 4, 0, 0, 0, 35, 9], [26, 4, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [39, 4, 1, 1, 1, 3, 4, 1, 0, 0, 40, 9], [32, 4, 1, 1, 1, 3, 4, 0, 0, 0, 2, 9], [28, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [47, 4, 5, 1, 6, 3, 4, 1, 0, 0, 60, 9], [25, 4, 1, 1, 4, 3, 2, 1, 0, 0, 40, 0]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
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
  storageUri: "gs://seldon-models/scv2/examples/income/classifier"
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
      "lastTransitionTime": "2022-11-24T11:55:46Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2022-11-24T11:55:46Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer income --inference-host ${MESH_IP}:80 \
```yaml
'{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```
```

```json
{
	"model_name": "income_1",
	"model_version": "1",
	"id": "c4beb76f-bf83-40a5-b320-812cac156e07",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": null,
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
```yaml
'{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```
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
  storageUri: "gs://seldon-models/scv2/examples/income/explainer"
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
      "lastTransitionTime": "2022-11-24T11:56:03Z",
      "status": "True",
      "type": "ModelReady"
    },
    {
      "lastTransitionTime": "2022-11-24T11:56:03Z",
      "status": "True",
      "type": "Ready"
    }
  ],
  "replicas": 1
}

```

```bash
seldon model infer income-explainer --inference-host ${MESH_IP}:80 \
```yaml
'{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```
```

```json
{
	"model_name": "income-explainer_1",
	"model_version": "1",
	"id": "dc39dca3-dcd1-431a-b86c-42c583f26ec5",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "explanation",
			"shape": [
				1,
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "str",
				"headers": null
			},
			"data": [
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9967845659163987, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.7874659400544959, 0.9083333333333333, 0.9967845659163987], \"precision\": [0.7874659400544959, 0.9083333333333333, 0.9967845659163987], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[45, 4, 1, 1, 5, 0, 4, 1, 0, 0, 50, 9], [54, 4, 1, 1, 4, 0, 4, 1, 0, 0, 48, 9], [48, 0, 1, 1, 0, 0, 4, 1, 0, 0, 55, 9], [40, 4, 5, 1, 5, 0, 4, 1, 0, 0, 40, 9], [55, 6, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [26, 4, 1, 1, 6, 1, 1, 1, 0, 0, 20, 4], [40, 4, 1, 1, 5, 0, 4, 1, 0, 0, 44, 9], [40, 4, 5, 1, 1, 3, 4, 0, 0, 0, 40, 9], [40, 7, 1, 1, 2, 1, 4, 1, 0, 0, 40, 9], [34, 4, 5, 1, 6, 0, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[53, 1, 1, 1, 8, 0, 4, 1, 0, 0, 48, 9], [36, 4, 1, 1, 5, 1, 4, 1, 14084, 0, 40, 9], [79, 0, 5, 1, 0, 0, 4, 1, 20051, 0, 40, 4], [42, 6, 1, 1, 6, 0, 4, 1, 0, 0, 50, 9], [33, 4, 5, 1, 5, 0, 4, 1, 15024, 0, 44, 9], [51, 2, 5, 1, 5, 0, 4, 1, 15024, 0, 40, 9], [28, 4, 1, 1, 5, 1, 4, 1, 0, 0, 55, 9], [51, 4, 1, 1, 1, 5, 4, 0, 7688, 0, 20, 9], [35, 4, 5, 1, 5, 0, 4, 1, 0, 1977, 45, 9], [41, 5, 5, 1, 8, 0, 4, 1, 0, 0, 60, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[33, 4, 5, 1, 8, 3, 4, 1, 7298, 0, 45, 9], [64, 5, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [39, 7, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [26, 4, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [45, 4, 5, 1, 1, 3, 1, 0, 0, 0, 40, 7], [35, 4, 1, 1, 8, 3, 4, 1, 0, 1848, 70, 9], [52, 4, 1, 1, 2, 3, 4, 1, 0, 0, 25, 9], [61, 0, 1, 1, 0, 3, 4, 0, 0, 0, 40, 9], [43, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [44, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[40, 6, 1, 1, 5, 3, 4, 1, 7298, 0, 50, 9], [52, 5, 2, 1, 5, 3, 4, 1, 99999, 0, 65, 9], [64, 4, 1, 1, 8, 3, 4, 1, 27828, 0, 50, 9], [49, 4, 1, 1, 5, 3, 4, 0, 99999, 0, 20, 9], [50, 4, 1, 1, 4, 3, 4, 1, 15024, 0, 40, 9], [28, 4, 1, 1, 5, 3, 4, 0, 0, 1564, 40, 9], [49, 4, 5, 1, 8, 3, 4, 1, 15024, 0, 50, 9], [50, 4, 1, 1, 5, 3, 4, 1, 15024, 0, 40, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[65, 0, 1, 1, 0, 3, 4, 1, 0, 0, 40, 9], [41, 4, 1, 1, 1, 3, 1, 1, 0, 0, 50, 9], [34, 4, 1, 1, 7, 3, 4, 1, 0, 0, 40, 8], [48, 5, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [25, 4, 1, 1, 4, 3, 2, 1, 0, 0, 40, 0], [59, 4, 5, 1, 5, 3, 4, 1, 0, 0, 50, 9], [30, 2, 5, 1, 5, 3, 4, 1, 0, 0, 40, 9], [51, 2, 5, 1, 5, 3, 4, 1, 0, 0, 40, 9], [41, 4, 5, 1, 5, 3, 4, 0, 0, 0, 40, 9], [34, 4, 5, 1, 8, 3, 4, 0, 0, 0, 50, 9]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
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

```python

```
