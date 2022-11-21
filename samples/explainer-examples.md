## Anchor Tabular Explainer for SKLearn Income Model


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
!seldon model load -f ./models/income.yaml
```

    {}



```python
!seldon model status income -w ModelAvailable
```

    {}



```python
!seldon model infer income \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "e41ac2cf-6c53-48a1-90cd-bbb4b3bb64d1",
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
!seldon model load -f ./models/income-explainer.yaml
```

    {}



```python
!seldon model status income-explainer -w ModelAvailable
```

    {}



```python
!seldon model infer income-explainer \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "1f50efc2-848a-42c7-85cf-978da74fab7d",
    	"parameters": {
    		"content_type": null,
    		"headers": null
    	},
    	"outputs": [
    		{
    			"name": "explanation",
    			"shape": [
    				1
    			],
    			"datatype": "BYTES",
    			"parameters": {
    				"content_type": "str",
    				"headers": null
    			},
    			"data": [
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9973684210526316, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.8037383177570093, 0.9179954441913439, 0.9973684210526316], \"precision\": [0.8037383177570093, 0.9179954441913439, 0.9973684210526316], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[32, 0, 1, 1, 0, 1, 4, 0, 0, 0, 20, 0], [40, 4, 1, 1, 6, 1, 4, 1, 0, 0, 53, 9], [26, 4, 5, 1, 2, 0, 4, 1, 0, 0, 20, 1], [42, 4, 1, 1, 6, 0, 4, 1, 0, 0, 40, 9], [43, 4, 1, 1, 2, 0, 4, 1, 0, 0, 48, 9], [37, 2, 1, 1, 7, 0, 4, 1, 0, 0, 40, 9], [34, 4, 1, 1, 2, 0, 4, 1, 0, 0, 50, 9], [57, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [55, 5, 5, 1, 5, 1, 4, 0, 0, 0, 35, 9], [44, 6, 5, 1, 5, 0, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[51, 5, 2, 1, 5, 0, 4, 1, 15024, 0, 40, 9], [39, 4, 5, 1, 5, 0, 4, 1, 0, 0, 50, 9], [34, 7, 5, 1, 5, 0, 4, 1, 7688, 0, 38, 9], [52, 4, 2, 1, 8, 0, 4, 1, 0, 0, 57, 9], [46, 4, 5, 1, 5, 0, 4, 1, 0, 0, 40, 9], [37, 4, 5, 1, 8, 0, 4, 1, 0, 0, 40, 9], [37, 6, 2, 1, 5, 1, 4, 0, 0, 2559, 60, 9], [52, 5, 2, 1, 5, 0, 4, 1, 99999, 0, 65, 9], [65, 5, 1, 1, 8, 0, 4, 1, 99999, 0, 65, 9], [70, 6, 1, 1, 5, 0, 4, 1, 20051, 0, 35, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[46, 7, 1, 1, 5, 3, 4, 1, 0, 0, 37, 9], [37, 2, 1, 1, 5, 3, 4, 0, 0, 0, 30, 9], [46, 4, 1, 1, 2, 3, 4, 1, 0, 0, 50, 0], [45, 4, 1, 1, 6, 3, 4, 1, 0, 2002, 55, 9], [41, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [45, 4, 1, 1, 5, 3, 4, 1, 0, 0, 30, 9], [39, 4, 1, 1, 1, 3, 4, 1, 0, 0, 40, 9], [29, 4, 1, 1, 5, 3, 4, 0, 0, 0, 45, 9], [34, 7, 2, 1, 4, 3, 1, 1, 0, 0, 20, 2], [31, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[30, 4, 1, 1, 6, 3, 4, 1, 7298, 0, 50, 9], [34, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [65, 4, 1, 1, 8, 3, 4, 1, 99999, 0, 40, 9], [90, 4, 1, 1, 6, 3, 4, 1, 9386, 0, 15, 9], [41, 4, 5, 1, 5, 3, 4, 1, 15024, 0, 50, 9], [46, 4, 5, 1, 5, 3, 4, 0, 7688, 0, 35, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[26, 4, 5, 1, 5, 3, 4, 1, 0, 0, 40, 9], [30, 7, 1, 1, 5, 3, 4, 1, 0, 0, 35, 9], [29, 4, 1, 1, 1, 3, 1, 0, 0, 0, 40, 6], [44, 6, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [47, 4, 1, 1, 1, 3, 4, 1, 0, 0, 40, 9], [49, 4, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [62, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [47, 4, 5, 1, 5, 3, 4, 0, 0, 0, 50, 9], [57, 4, 1, 1, 6, 3, 4, 1, 0, 0, 55, 9], [37, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
    			]
    		}
    	]
    }



```python
!seldon model unload income-explainer
```

    {}



```python
!seldon model unload income
```

    {}



```python

```
