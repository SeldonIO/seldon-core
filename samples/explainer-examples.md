## Anchor Tabular Explainer for SKLearn Income Model

```bash
cat ./models/income.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.1/income/classifier"
  requirements:
  - sklearn

```

```bash
seldon model load -f ./models/income.yaml
```

```json
{}

```

```bash
seldon model status income -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer income \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income_1",
	"model_version": "1",
	"id": "f1cadfd0-24b8-4dc8-aede-cf3c1dd0973c",
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
seldon model load -f ./models/income-explainer.yaml
```

```json
{}

```

```bash
seldon model status income-explainer -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer income-explainer \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income-explainer_1",
	"model_version": "1",
	"id": "67460faa-a429-464f-8108-269ae21c024d",
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
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Relationship = Own-child\", \"Marital Status = Never-Married\"], \"precision\": 0.9666666666666667, \"coverage\": 0.07165109034267912, \"raw\": {\"feature\": [5, 3], \"mean\": [0.8118811881188119, 0.9666666666666667], \"precision\": [0.8118811881188119, 0.9666666666666667], \"coverage\": [0.0787716955941255, 0.07165109034267912], \"examples\": [{\"covered_true\": [[26, 4, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [47, 4, 1, 0, 7, 3, 4, 1, 0, 0, 40, 8], [25, 4, 1, 1, 7, 3, 4, 0, 0, 0, 20, 9], [26, 2, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [38, 4, 1, 2, 8, 3, 4, 1, 0, 0, 55, 9], [26, 4, 1, 0, 8, 3, 4, 0, 0, 0, 55, 9], [52, 4, 1, 0, 8, 3, 4, 1, 0, 0, 55, 9], [49, 5, 1, 2, 5, 3, 4, 1, 6497, 0, 45, 9], [29, 0, 1, 1, 0, 3, 4, 1, 0, 0, 40, 9], [45, 5, 1, 0, 8, 3, 4, 1, 0, 0, 45, 9]], \"covered_false\": [[74, 6, 1, 3, 2, 3, 4, 1, 15831, 0, 8, 3], [45, 2, 1, 0, 2, 3, 1, 1, 7298, 0, 40, 9], [34, 7, 2, 0, 5, 3, 4, 0, 0, 0, 50, 9], [59, 4, 5, 0, 8, 3, 4, 1, 0, 0, 40, 9], [42, 5, 1, 0, 6, 3, 4, 1, 0, 0, 60, 9], [44, 4, 1, 0, 5, 3, 4, 1, 7688, 0, 40, 9], [46, 5, 1, 0, 6, 3, 4, 1, 0, 1902, 42, 0], [36, 4, 1, 0, 6, 3, 4, 1, 0, 2415, 45, 9], [41, 6, 1, 0, 2, 3, 4, 1, 0, 0, 75, 9], [48, 4, 5, 0, 8, 3, 4, 1, 0, 0, 40, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[48, 4, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [36, 6, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [31, 6, 1, 1, 8, 3, 4, 1, 0, 0, 60, 9], [47, 4, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [27, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [44, 6, 1, 1, 8, 3, 1, 1, 0, 0, 48, 4], [44, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [24, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [49, 4, 1, 1, 2, 3, 4, 1, 0, 1902, 40, 9], [35, 4, 5, 1, 8, 3, 1, 1, 0, 0, 65, 2]], \"covered_false\": [[38, 4, 1, 1, 6, 3, 4, 1, 7298, 0, 40, 9], [41, 6, 5, 1, 8, 3, 4, 1, 7298, 0, 70, 9], [42, 2, 1, 1, 1, 3, 4, 0, 99999, 0, 40, 9], [63, 6, 1, 1, 6, 3, 4, 1, 10605, 0, 40, 9], [44, 5, 1, 1, 8, 3, 4, 1, 99999, 0, 45, 9], [41, 4, 1, 1, 8, 3, 4, 1, 7298, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Relationship = Own-child\", \"Marital Status = Never-Married\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
			]
		}
	]
}

```

```bash
seldon model unload income-explainer
```

```json
{}

```

```bash
seldon model unload income
```

```json
{}

```

## Anchor Text Explainer for SKLearn Movies Sentiment Model

```bash
cat ./models/moviesentiment.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: sentiment
spec:
  storageUri: "gs://seldon-models/scv2/examples/moviesentiment/classifier"
  requirements:
  - sklearn

```

```bash
seldon model load -f ./models/moviesentiment.yaml
```

```json
{}

```

```bash
seldon model status sentiment -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer sentiment \
  '{"parameters": {"content_type": "str"}, "inputs": [{"name": "foo", "data": ["I am good"], "datatype": "BYTES","shape": [1]}]}'
```

```json
{
	"model_name": "sentiment_1",
	"model_version": "1",
	"id": "6c514a89-2859-4835-97b0-a73336233726",
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
				1
			]
		}
	]
}

```

```bash
cat ./models/moviesentiment-explainer.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: sentiment-explainer
spec:
  storageUri: "gs://seldon-models/scv2/examples/moviesentiment/explainer"
  explainer:
    type: anchor_text
    modelRef: sentiment

```

```bash
seldon model load -f ./models/moviesentiment-explainer.yaml
```

```json
{}

```

```bash
seldon model status sentiment-explainer -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer sentiment-explainer \
  '{"parameters": {"content_type": "str"}, "inputs": [{"name": "foo", "data": ["I am good"], "datatype": "BYTES","shape": [1]}]}'
```

```json
{
	"model_name": "sentiment-explainer_1",
	"model_version": "1",
	"id": "1d5859a9-9d1b-4edc-b1c2-b40f8369c804",
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
				"{\"meta\": {\"name\": \"AnchorText\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 0, \"sample_proba\": 0.5}, \"version\": \"0.9.0\"}, \"data\": {\"anchor\": [\"good\"], \"precision\": 1.0, \"coverage\": 0.5006, \"raw\": {\"feature\": [2], \"mean\": [1.0], \"precision\": [1.0], \"coverage\": [0.5006], \"examples\": [{\"covered_true\": [\"I UNK good\", \"I UNK good\", \"I am good\", \"I UNK good\", \"UNK am good\", \"I UNK good\", \"UNK UNK good\", \"I UNK good\", \"I UNK good\", \"UNK am good\"], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"good\"], \"positions\": [5], \"instance\": \"I am good\", \"instances\": [\"I am good\"], \"prediction\": [1]}}}"
			]
		}
	]
}

```

```bash
seldon model unload sentiment-explainer
```

```json
{}

```

```bash
seldon model unload sentiment
```

```json
{}

```

```python

```
