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
    	"id": "6278d5d2-022f-4b47-aca3-4387fe63dbb7",
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
    	"id": "247ddf15-8e21-49f4-bbbc-8dc7d312d895",
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
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9977973568281938, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.8459214501510574, 0.9085714285714286, 0.9977973568281938], \"precision\": [0.8459214501510574, 0.9085714285714286, 0.9977973568281938], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[28, 2, 1, 1, 5, 1, 2, 0, 0, 0, 60, 9], [26, 0, 1, 1, 0, 1, 4, 0, 0, 0, 80, 9], [49, 4, 5, 1, 6, 0, 4, 1, 0, 0, 40, 9], [26, 4, 1, 1, 6, 0, 4, 1, 0, 0, 48, 9], [50, 4, 5, 1, 2, 0, 4, 1, 0, 0, 30, 9], [46, 6, 1, 1, 2, 0, 4, 1, 0, 0, 40, 9], [40, 4, 1, 1, 1, 0, 4, 1, 0, 0, 50, 9], [42, 4, 5, 1, 6, 0, 4, 1, 0, 0, 40, 8], [52, 5, 1, 1, 8, 0, 4, 1, 0, 0, 60, 9], [67, 0, 5, 1, 0, 1, 4, 0, 0, 0, 40, 9]], \"covered_false\": [[40, 1, 1, 1, 1, 0, 4, 1, 7298, 0, 48, 9], [40, 4, 1, 1, 7, 5, 1, 0, 7688, 0, 52, 6], [26, 4, 1, 1, 8, 1, 4, 1, 0, 0, 70, 9], [45, 4, 5, 1, 8, 0, 4, 1, 15024, 0, 60, 9], [42, 4, 1, 1, 6, 1, 4, 1, 10520, 0, 50, 9], [42, 6, 1, 1, 2, 0, 4, 1, 15024, 0, 60, 9], [48, 1, 1, 1, 1, 5, 4, 0, 7688, 0, 40, 9], [32, 4, 1, 1, 8, 0, 4, 1, 7688, 0, 40, 9], [37, 4, 5, 1, 8, 0, 4, 1, 0, 0, 40, 9], [37, 4, 1, 1, 8, 0, 4, 1, 15024, 0, 40, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[33, 4, 1, 1, 8, 3, 1, 0, 0, 0, 40, 7], [22, 4, 1, 1, 2, 3, 4, 0, 0, 0, 15, 9], [36, 4, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [43, 4, 1, 1, 7, 3, 4, 0, 0, 0, 35, 0], [25, 0, 1, 1, 0, 3, 4, 1, 0, 0, 40, 1], [40, 7, 2, 1, 5, 3, 4, 1, 0, 1887, 50, 9], [49, 4, 1, 1, 1, 3, 4, 1, 5178, 0, 40, 9], [46, 4, 1, 1, 2, 3, 4, 1, 0, 0, 48, 9], [47, 6, 1, 1, 8, 3, 4, 1, 0, 0, 60, 9], [36, 7, 1, 1, 5, 3, 4, 1, 0, 1876, 44, 9]], \"covered_false\": [[39, 4, 1, 1, 8, 3, 4, 1, 15024, 0, 45, 9], [39, 4, 5, 1, 5, 3, 4, 1, 99999, 0, 40, 9], [27, 4, 1, 1, 5, 3, 1, 1, 13550, 0, 40, 9], [51, 4, 5, 1, 5, 3, 4, 1, 7688, 0, 45, 9], [23, 6, 1, 1, 1, 3, 4, 1, 0, 2231, 40, 9], [56, 5, 1, 1, 8, 3, 4, 1, 15024, 0, 50, 9], [41, 4, 1, 1, 6, 3, 4, 1, 15024, 0, 45, 9], [50, 4, 1, 1, 4, 3, 4, 1, 15024, 0, 40, 9], [50, 4, 1, 1, 8, 3, 4, 1, 7298, 0, 50, 9], [38, 4, 1, 1, 5, 3, 4, 0, 7688, 0, 40, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[46, 4, 5, 1, 8, 3, 4, 0, 0, 0, 50, 9], [40, 6, 5, 1, 8, 3, 4, 0, 0, 0, 20, 9], [23, 2, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [37, 4, 1, 1, 4, 3, 4, 0, 0, 0, 40, 9], [43, 6, 1, 1, 2, 3, 4, 1, 0, 0, 35, 9], [43, 6, 1, 1, 2, 3, 1, 0, 0, 0, 80, 7], [48, 4, 1, 1, 8, 3, 2, 1, 0, 0, 40, 9], [27, 7, 1, 1, 5, 3, 2, 1, 0, 0, 40, 9], [46, 4, 1, 1, 8, 3, 4, 1, 0, 0, 55, 9], [26, 2, 1, 1, 1, 3, 4, 1, 0, 0, 41, 9]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
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
