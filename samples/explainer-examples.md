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
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.3/income/classifier"
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
	"id": "e06e7238-d25f-412c-a3c4-ea77721d660d",
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
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.3/income/explainer"
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
	"id": "b8c18f49-3944-4841-8d1d-81585ba40830",
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
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.9.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 1.0, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.8063492063492064, 0.9221902017291066, 1.0], \"precision\": [0.8063492063492064, 0.9221902017291066, 1.0], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[44, 7, 1, 1, 5, 0, 0, 1, 0, 0, 38, 9], [74, 6, 1, 1, 8, 0, 4, 1, 0, 1825, 12, 9], [58, 4, 1, 1, 2, 3, 4, 0, 0, 0, 20, 9], [48, 4, 1, 1, 6, 4, 4, 1, 0, 0, 40, 9], [22, 4, 1, 1, 2, 3, 4, 0, 0, 0, 15, 9], [57, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [36, 4, 5, 1, 5, 0, 4, 1, 0, 0, 40, 9], [28, 7, 1, 1, 5, 3, 4, 1, 0, 0, 16, 9], [49, 7, 1, 1, 1, 0, 4, 1, 0, 0, 40, 9], [27, 4, 1, 1, 4, 1, 4, 0, 0, 0, 40, 9]], \"covered_false\": [[50, 5, 1, 1, 8, 0, 4, 1, 15024, 0, 60, 9], [56, 4, 1, 1, 1, 0, 4, 1, 5178, 0, 44, 9], [50, 5, 2, 1, 5, 0, 4, 1, 15024, 0, 60, 9], [66, 4, 1, 1, 8, 0, 4, 1, 99999, 0, 55, 0], [36, 4, 1, 1, 5, 0, 4, 1, 15024, 0, 50, 9], [59, 4, 5, 1, 8, 0, 4, 1, 0, 0, 45, 9], [36, 4, 1, 1, 6, 0, 4, 1, 0, 1902, 45, 9], [37, 1, 1, 1, 1, 0, 1, 1, 7298, 0, 40, 7], [34, 4, 1, 1, 6, 0, 4, 1, 7298, 0, 60, 9], [52, 2, 5, 1, 5, 1, 4, 1, 0, 0, 32, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[33, 4, 1, 1, 8, 3, 4, 1, 0, 0, 50, 9], [76, 4, 1, 1, 5, 3, 4, 0, 0, 0, 20, 9], [46, 4, 1, 1, 8, 3, 4, 1, 0, 0, 53, 9], [31, 5, 1, 1, 8, 3, 4, 1, 0, 0, 55, 9], [39, 4, 1, 1, 6, 3, 4, 1, 3464, 0, 40, 9], [41, 2, 1, 1, 5, 3, 4, 1, 0, 0, 20, 9], [30, 2, 1, 1, 5, 3, 4, 0, 0, 0, 15, 9], [37, 2, 1, 1, 5, 3, 4, 0, 0, 0, 30, 9], [45, 4, 1, 1, 5, 3, 4, 1, 0, 0, 55, 9], [54, 6, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[50, 4, 5, 1, 6, 3, 4, 1, 15024, 0, 40, 9], [45, 4, 1, 1, 6, 3, 4, 1, 8614, 0, 48, 9], [44, 4, 5, 1, 8, 3, 4, 1, 14084, 0, 56, 9], [39, 5, 1, 1, 8, 3, 4, 1, 7298, 0, 40, 9], [40, 4, 1, 1, 7, 3, 1, 0, 7688, 0, 52, 6], [64, 4, 1, 1, 8, 3, 4, 1, 27828, 0, 55, 9], [62, 5, 1, 1, 6, 3, 4, 1, 99999, 0, 40, 9], [55, 4, 1, 1, 8, 3, 4, 1, 15024, 0, 48, 9], [32, 5, 1, 1, 2, 3, 4, 1, 7688, 0, 50, 9], [39, 4, 1, 1, 8, 3, 4, 1, 99999, 0, 70, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[66, 5, 1, 1, 8, 3, 4, 1, 0, 0, 70, 9], [49, 4, 1, 1, 8, 3, 2, 0, 0, 0, 60, 9], [40, 4, 2, 1, 5, 3, 4, 1, 0, 0, 40, 9], [59, 0, 1, 1, 0, 3, 4, 1, 0, 0, 40, 9], [62, 4, 5, 1, 5, 3, 4, 0, 0, 0, 45, 9], [40, 4, 1, 1, 8, 3, 4, 1, 0, 1902, 32, 9], [37, 4, 1, 1, 7, 3, 4, 1, 0, 0, 70, 1], [38, 2, 5, 1, 8, 3, 4, 1, 0, 0, 70, 9], [23, 4, 1, 1, 1, 3, 4, 0, 0, 0, 30, 9], [32, 4, 1, 1, 6, 3, 2, 1, 0, 0, 50, 9]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
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
