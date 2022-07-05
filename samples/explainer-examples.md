## Anchor Tabular Explainer for SKLearn Income Model


```bash
cat ./models/income.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/income"
      requirements:
      - sklearn
```
````

```bash
seldon model load -f ./models/income.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model status income -w ModelAvailable
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model infer income \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[53,4,0,2,8,4,2,0,0,0,60,9]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "07975062-e883-43b5-882f-524bfd380806",
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
```
````

```bash
cat ./models/income-explainer.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/alibi-explain/income"
      explainer:
        type: anchor_tabular
        modelRef: income
```
````

```bash
seldon model load -f ./models/income-explainer.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model status income-explainer -w ModelAvailable
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model infer income-explainer \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[53,4,0,2,8,4,2,0,0,0,60,9]]}]}'
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "12eef6f5-8d87-420e-8c3f-4091014be7ce",
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
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.7.0\"}, \"data\": {\"anchor\": [\"Marital Status = Separated\", \"Capital Gain <= 0.00\"], \"precision\": 0.9829351535836177, \"coverage\": 0.16943333333333332, \"raw\": {\"feature\": [3, 8], \"mean\": [0.9465408805031447, 0.9829351535836177], \"precision\": [0.9465408805031447, 0.9829351535836177], \"coverage\": [0.18063333333333334, 0.16943333333333332], \"examples\": [{\"covered_true\": [[19, 4, 3, 2, 2, 3, 4, 1, 0, 0, 40, 9], [27, 4, 4, 2, 6, 0, 4, 1, 0, 0, 40, 9], [39, 4, 4, 2, 2, 0, 4, 1, 0, 0, 40, 9], [27, 4, 4, 2, 2, 0, 4, 1, 0, 0, 45, 9], [32, 0, 3, 2, 0, 1, 4, 0, 0, 0, 49, 9], [27, 4, 4, 2, 1, 1, 4, 1, 4416, 0, 40, 9], [58, 5, 1, 2, 5, 0, 4, 1, 0, 0, 40, 9], [47, 4, 4, 2, 1, 4, 4, 0, 0, 0, 20, 9], [24, 4, 4, 2, 7, 3, 4, 0, 0, 0, 30, 9], [30, 6, 4, 2, 2, 1, 4, 1, 0, 0, 35, 9]], \"covered_false\": [[33, 4, 5, 2, 8, 0, 4, 1, 0, 1902, 45, 3], [34, 4, 1, 2, 8, 1, 4, 1, 0, 0, 85, 1], [64, 5, 1, 2, 8, 0, 4, 1, 15024, 0, 55, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[31, 4, 4, 2, 7, 1, 4, 0, 0, 0, 40, 9], [55, 4, 3, 2, 6, 3, 4, 0, 0, 0, 20, 9], [39, 4, 0, 2, 2, 3, 4, 1, 0, 0, 40, 9], [36, 4, 0, 2, 5, 5, 4, 0, 0, 0, 40, 9], [49, 6, 4, 2, 2, 1, 4, 1, 0, 0, 55, 9], [25, 4, 4, 2, 7, 3, 4, 0, 0, 0, 8, 9], [39, 4, 4, 2, 7, 5, 4, 0, 0, 0, 40, 9], [65, 4, 4, 2, 6, 4, 4, 0, 0, 0, 25, 0], [54, 6, 5, 2, 5, 1, 4, 1, 0, 0, 50, 9], [36, 4, 0, 2, 1, 4, 4, 0, 0, 0, 40, 9]], \"covered_false\": [[48, 4, 5, 2, 5, 0, 4, 1, 0, 1902, 40, 9], [43, 4, 5, 2, 8, 4, 4, 0, 0, 2547, 40, 9], [43, 6, 1, 2, 5, 5, 4, 0, 0, 1887, 70, 9], [41, 4, 6, 2, 5, 0, 4, 1, 0, 2415, 40, 9], [44, 4, 1, 2, 8, 0, 4, 1, 0, 1902, 56, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Separated\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [53.0, 4.0, 0.0, 2.0, 8.0, 4.0, 2.0, 0.0, 0.0, 0.0, 60.0, 9.0], \"instances\": [[53.0, 4.0, 0.0, 2.0, 8.0, 4.0, 2.0, 0.0, 0.0, 0.0, 60.0, 9.0]]}}}"
    			]
    		}
    	]
    }
```
````

```bash
seldon model unload income-explainer
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload income
```
````{collapse} Expand to see output
```json

    {}
```
````

```python

```
