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




    '172.19.255.1'



## Explain Model


```python
!cat ./models/income.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/classifier"
      requirements:
      - sklearn



```python
!kubectl create -f ./models/income.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/income created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/income condition met



```python
!kubectl get model income -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-24T11:18:07Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-11-24T11:18:07Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }



```python
!seldon model infer income --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "fb46ef60-9b15-4999-80b8-3d5735f1930a",
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



```python
!cat ./models/income-explainer.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/explainer"
      explainer:
        type: anchor_tabular
        modelRef: income



```python
!kubectl create -f ./models/income-explainer.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/income-explainer created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/income condition met
    model.mlops.seldon.io/income-explainer condition met



```python
!kubectl get model income-explainer -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-24T11:18:31Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-11-24T11:18:31Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }



```python
!seldon model infer income-explainer --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "cf387a7e-8d38-4580-be3c-c74947d5010a",
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
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9939759036144579, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.7870036101083032, 0.9122340425531915, 0.9939759036144579], \"precision\": [0.7870036101083032, 0.9122340425531915, 0.9939759036144579], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[33, 1, 5, 1, 2, 0, 2, 1, 0, 0, 40, 5], [27, 4, 1, 1, 7, 1, 4, 0, 0, 0, 40, 9], [32, 4, 1, 1, 5, 1, 4, 1, 0, 0, 38, 9], [44, 4, 5, 1, 8, 0, 4, 1, 5178, 0, 40, 9], [24, 4, 5, 1, 5, 1, 4, 0, 0, 0, 40, 9], [38, 2, 1, 1, 5, 5, 4, 0, 0, 0, 20, 9], [65, 0, 1, 1, 0, 0, 4, 1, 0, 0, 40, 9], [42, 6, 1, 1, 6, 0, 4, 1, 0, 0, 45, 9], [37, 6, 1, 1, 6, 5, 4, 0, 0, 0, 50, 9], [44, 4, 1, 1, 8, 0, 3, 1, 0, 0, 40, 5]], \"covered_false\": [[40, 2, 1, 1, 5, 1, 4, 1, 0, 0, 50, 9], [44, 6, 1, 1, 8, 0, 4, 1, 0, 2415, 50, 9], [29, 4, 1, 1, 5, 1, 4, 1, 0, 0, 50, 9], [46, 4, 1, 1, 8, 0, 4, 1, 0, 0, 60, 9], [59, 4, 2, 1, 8, 0, 4, 1, 0, 0, 40, 9], [42, 4, 1, 1, 6, 1, 4, 1, 10520, 0, 50, 9], [45, 4, 2, 1, 5, 4, 2, 0, 0, 3004, 35, 9], [53, 4, 5, 1, 8, 0, 4, 1, 15024, 0, 55, 9], [66, 2, 2, 1, 5, 0, 2, 1, 20051, 0, 35, 5], [64, 6, 2, 1, 8, 0, 4, 1, 0, 0, 25, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[33, 4, 5, 1, 5, 3, 2, 1, 0, 0, 50, 9], [33, 4, 1, 1, 1, 3, 1, 1, 0, 0, 40, 1], [55, 2, 5, 1, 5, 3, 4, 0, 0, 0, 45, 9], [43, 2, 5, 1, 5, 3, 2, 0, 0, 0, 37, 9], [41, 5, 1, 1, 8, 3, 1, 1, 0, 1977, 60, 2], [36, 4, 1, 1, 6, 3, 4, 1, 0, 1887, 35, 9], [23, 4, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [33, 6, 1, 1, 6, 3, 4, 0, 0, 0, 40, 9], [26, 2, 1, 1, 5, 3, 2, 1, 0, 0, 38, 1], [25, 4, 1, 1, 8, 3, 4, 1, 0, 0, 55, 9]], \"covered_false\": [[42, 1, 5, 1, 8, 3, 4, 0, 14084, 0, 60, 9], [45, 4, 2, 1, 5, 3, 4, 1, 15020, 0, 40, 6], [37, 4, 1, 1, 8, 3, 4, 0, 7688, 0, 45, 9], [51, 5, 1, 1, 8, 3, 4, 1, 15024, 0, 40, 9], [49, 7, 2, 1, 8, 3, 4, 0, 0, 2258, 50, 9], [37, 4, 5, 1, 8, 3, 4, 1, 27828, 0, 60, 6], [48, 4, 1, 1, 5, 3, 4, 1, 15024, 0, 45, 9], [45, 6, 2, 1, 5, 3, 4, 1, 15024, 0, 40, 9], [31, 4, 1, 1, 1, 3, 4, 1, 8614, 0, 40, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[25, 2, 1, 1, 5, 3, 4, 1, 0, 0, 45, 9], [49, 5, 1, 1, 6, 3, 4, 1, 0, 0, 50, 9], [42, 1, 1, 1, 8, 3, 4, 1, 0, 0, 52, 9], [47, 4, 1, 1, 8, 3, 4, 1, 0, 0, 50, 9], [24, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [30, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [36, 6, 1, 1, 8, 3, 4, 1, 0, 0, 50, 6], [42, 4, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [40, 5, 1, 1, 5, 3, 4, 1, 0, 0, 60, 9], [35, 1, 1, 1, 5, 3, 1, 1, 0, 0, 40, 7]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
    			]
    		}
    	]
    }



```python
!kubectl delete -f ./models/income.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/income-explainer.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "income" deleted
    model.mlops.seldon.io "income-explainer" deleted


## Explain Pipeline


```python
!cat ./models/income.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/classifier"
      requirements:
      - sklearn



```python
!kubectl create -f ./models/income.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/income created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/income condition met



```python
!kubectl get model income -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-24T11:19:17Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-11-24T11:19:17Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }



```python
!seldon model infer income --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "6fb781d1-5117-464b-b1b0-5fdba6bd93ed",
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



```python
!cat ./pipelines/income-v1.yaml
```

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
    



```python
!kubectl create -f ./pipelines/income-v1.yaml -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/income-prod created



```python
!kubectl wait --for condition=ready --timeout=300s pipeline --all -n ${NAMESPACE}
```

    pipeline.mlops.seldon.io/income-prod condition met



```python
!seldon pipeline infer income-prod --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

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



```python
!cat ./models/income-explainer-pipeline.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/explainer"
      explainer:
        type: anchor_tabular
        pipelineRef: income-prod



```python
!kubectl create -f ./models/income-explainer-pipeline.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/income-explainer created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
```

    model.mlops.seldon.io/income condition met
    model.mlops.seldon.io/income-explainer condition met



```python
!kubectl get model income-explainer -n ${NAMESPACE} -o jsonpath='{.status}' | jq -M .
```

    {
      "conditions": [
        {
          "lastTransitionTime": "2022-11-24T11:20:24Z",
          "status": "True",
          "type": "ModelReady"
        },
        {
          "lastTransitionTime": "2022-11-24T11:20:24Z",
          "status": "True",
          "type": "Ready"
        }
      ],
      "replicas": 1
    }



```python
!seldon model infer income-explainer --inference-host ${MESH_IP}:80 \
     '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "d10fd321-d553-4ba1-be83-dd5a9a903448",
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
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9966216216216216, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.8258823529411765, 0.9281437125748503, 0.9966216216216216], \"precision\": [0.8258823529411765, 0.9281437125748503, 0.9966216216216216], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[52, 4, 5, 1, 4, 0, 4, 1, 0, 0, 45, 9], [65, 4, 1, 1, 2, 0, 4, 1, 0, 0, 40, 1], [61, 5, 5, 1, 8, 4, 4, 0, 0, 0, 99, 9], [38, 4, 5, 1, 8, 0, 1, 1, 0, 0, 40, 0], [48, 4, 5, 1, 8, 0, 4, 1, 0, 0, 40, 9], [30, 2, 1, 1, 5, 4, 4, 0, 0, 0, 40, 9], [50, 4, 1, 1, 6, 0, 4, 1, 0, 0, 40, 9], [38, 4, 1, 1, 2, 1, 4, 1, 0, 0, 40, 9], [30, 4, 1, 1, 6, 0, 4, 1, 2407, 0, 40, 9], [62, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[49, 2, 5, 1, 8, 0, 2, 1, 0, 0, 47, 9], [53, 4, 1, 1, 8, 0, 4, 1, 0, 0, 45, 9], [49, 4, 1, 1, 8, 1, 2, 0, 27828, 0, 60, 9], [47, 2, 5, 1, 5, 0, 4, 1, 7688, 0, 60, 9], [38, 4, 5, 1, 8, 0, 4, 1, 0, 0, 60, 9], [47, 4, 1, 1, 6, 0, 4, 1, 0, 0, 60, 9], [45, 4, 2, 1, 8, 4, 4, 1, 0, 0, 40, 9], [68, 7, 5, 1, 5, 0, 4, 1, 0, 2377, 42, 0], [57, 4, 1, 1, 8, 1, 4, 0, 0, 0, 45, 9], [50, 4, 5, 1, 5, 0, 4, 1, 15024, 0, 45, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[28, 4, 1, 1, 2, 3, 4, 1, 0, 0, 45, 9], [57, 1, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [63, 2, 5, 1, 5, 3, 4, 1, 0, 0, 48, 9], [45, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [45, 1, 1, 1, 8, 3, 4, 1, 0, 0, 50, 9], [50, 4, 5, 1, 5, 3, 4, 0, 0, 0, 35, 9], [51, 4, 2, 1, 5, 3, 1, 1, 0, 0, 99, 4], [29, 4, 1, 1, 5, 3, 4, 0, 0, 0, 46, 9], [45, 4, 1, 1, 1, 3, 2, 0, 0, 0, 40, 9], [25, 4, 1, 1, 8, 3, 4, 0, 3325, 0, 40, 9]], \"covered_false\": [[43, 4, 1, 1, 5, 3, 4, 0, 15024, 0, 50, 9], [43, 4, 1, 1, 5, 3, 4, 0, 14344, 0, 40, 9], [36, 4, 1, 1, 5, 3, 4, 1, 15024, 0, 50, 9], [47, 7, 2, 1, 5, 3, 4, 1, 15024, 0, 50, 9], [46, 4, 2, 1, 5, 3, 4, 0, 25236, 0, 65, 9], [71, 4, 1, 1, 2, 3, 4, 1, 11678, 0, 45, 9], [46, 7, 2, 1, 5, 3, 4, 1, 7688, 0, 45, 9], [54, 6, 1, 1, 2, 3, 4, 1, 27828, 0, 50, 9], [51, 4, 1, 1, 1, 3, 4, 0, 7688, 0, 20, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[26, 4, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [38, 0, 5, 1, 0, 3, 2, 0, 0, 0, 2, 9], [53, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [43, 7, 5, 1, 5, 3, 4, 1, 0, 1887, 45, 9], [28, 7, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [50, 1, 1, 1, 8, 3, 4, 1, 0, 0, 55, 9], [60, 0, 1, 1, 0, 3, 1, 1, 0, 2163, 25, 4], [54, 6, 5, 1, 6, 3, 4, 1, 0, 0, 60, 9], [51, 4, 1, 1, 8, 3, 4, 1, 0, 0, 60, 9], [25, 4, 1, 1, 8, 3, 4, 1, 0, 0, 45, 9]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
    			]
    		}
    	]
    }



```python
!kubectl delete -f ./models/income.yaml -n ${NAMESPACE}
!kubectl delete -f ./models/income-explainer.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "income" deleted
    model.mlops.seldon.io "income-explainer" deleted



```python

```
