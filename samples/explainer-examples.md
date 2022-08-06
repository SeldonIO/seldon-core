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
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/classifier"
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
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}' 
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income_1",
    	"model_version": "1",
    	"id": "f75d5245-32cd-4e1c-91ad-1eb8b23e579d",
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
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/explainer"
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
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```
````{collapse} Expand to see output
```json

    {
    	"model_name": "income-explainer_1",
    	"model_version": "1",
    	"id": "826ee324-758e-4a3d-8c1b-a0fc1c9af51b",
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
    				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.7.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9921875, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.7870662460567823, 0.8948545861297539, 0.9921875], \"precision\": [0.7870662460567823, 0.8948545861297539, 0.9921875], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[51, 1, 1, 1, 5, 1, 2, 0, 0, 0, 40, 9], [26, 7, 1, 1, 5, 1, 4, 1, 0, 0, 40, 9], [26, 4, 1, 1, 6, 0, 4, 1, 0, 0, 55, 9], [23, 4, 1, 1, 8, 1, 4, 0, 0, 0, 45, 9], [36, 6, 1, 1, 8, 0, 4, 1, 2407, 0, 40, 9], [36, 4, 1, 1, 6, 1, 4, 0, 0, 1741, 40, 9], [30, 2, 1, 1, 4, 0, 4, 1, 0, 0, 45, 9], [46, 4, 1, 1, 5, 0, 4, 1, 0, 1902, 40, 9], [46, 4, 1, 1, 8, 0, 4, 1, 0, 0, 35, 9], [65, 4, 1, 1, 2, 0, 4, 1, 0, 0, 20, 9]], \"covered_false\": [[35, 4, 5, 1, 8, 0, 4, 1, 0, 0, 50, 9], [36, 6, 2, 1, 5, 1, 4, 1, 0, 0, 40, 9], [46, 7, 1, 1, 8, 0, 4, 1, 7688, 0, 40, 9], [49, 7, 5, 1, 6, 0, 4, 1, 99999, 0, 80, 9], [52, 4, 5, 1, 5, 0, 4, 1, 0, 0, 40, 9], [29, 4, 5, 1, 6, 1, 4, 1, 0, 0, 50, 9], [68, 4, 5, 1, 8, 0, 4, 1, 0, 2392, 40, 9], [44, 7, 5, 1, 8, 0, 4, 1, 0, 1902, 40, 9], [37, 4, 5, 1, 8, 1, 4, 1, 0, 0, 45, 9], [39, 4, 1, 1, 8, 0, 4, 1, 15024, 0, 45, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[48, 4, 5, 1, 5, 3, 4, 1, 0, 0, 40, 6], [36, 2, 1, 1, 5, 3, 4, 0, 0, 0, 60, 9], [49, 4, 5, 1, 8, 3, 4, 1, 0, 1977, 40, 9], [75, 0, 2, 1, 0, 3, 4, 1, 0, 0, 40, 9], [53, 4, 1, 1, 6, 3, 4, 1, 0, 0, 50, 9], [36, 5, 1, 1, 6, 3, 4, 1, 0, 0, 60, 9], [49, 4, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [45, 4, 1, 1, 8, 3, 4, 1, 0, 0, 45, 9], [36, 4, 1, 1, 8, 3, 4, 1, 0, 0, 50, 9], [29, 4, 5, 1, 8, 3, 1, 1, 0, 0, 40, 6]], \"covered_false\": [[46, 5, 1, 1, 2, 3, 4, 1, 7298, 0, 40, 0], [30, 4, 1, 1, 1, 3, 4, 1, 7298, 0, 40, 9], [47, 4, 5, 1, 8, 3, 4, 1, 15024, 0, 55, 9], [44, 4, 1, 1, 5, 3, 4, 1, 15024, 0, 50, 3], [63, 1, 2, 1, 8, 3, 4, 0, 0, 2559, 60, 9], [44, 4, 5, 1, 2, 3, 4, 1, 15024, 0, 45, 9], [46, 4, 1, 1, 8, 3, 4, 1, 7688, 0, 40, 9], [51, 5, 2, 1, 5, 3, 4, 1, 15024, 0, 40, 9], [38, 6, 1, 1, 6, 3, 4, 1, 7298, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[57, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 4], [42, 4, 1, 1, 5, 3, 4, 0, 0, 0, 55, 9], [27, 2, 1, 1, 4, 3, 4, 1, 0, 0, 68, 9], [41, 4, 5, 1, 5, 3, 4, 0, 0, 0, 45, 1], [51, 1, 1, 1, 8, 3, 4, 1, 0, 1902, 40, 9], [49, 6, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [47, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [22, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [59, 7, 1, 1, 8, 3, 2, 0, 0, 0, 40, 5], [40, 4, 1, 1, 1, 3, 4, 0, 0, 0, 35, 9]], \"covered_false\": [[67, 5, 1, 1, 8, 3, 4, 1, 0, 2392, 75, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
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
