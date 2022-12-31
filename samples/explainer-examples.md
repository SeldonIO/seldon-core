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
  storageUri: "gs://seldon-models/scv2/examples/income/classifier"
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
	"id": "d53b1af1-5b06-460b-a764-d4e5f8a5fbf8",
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
	"id": "cf5d42c1-23b1-4645-9263-e033de07c55a",
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
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.8.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"precision\": 0.9879807692307693, \"coverage\": 0.06853582554517133, \"raw\": {\"feature\": [3, 5, 8], \"mean\": [0.8005502063273727, 0.907725321888412, 0.9879807692307693], \"precision\": [0.8005502063273727, 0.907725321888412, 0.9879807692307693], \"coverage\": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], \"examples\": [{\"covered_true\": [[28, 4, 1, 1, 1, 1, 4, 0, 0, 0, 45, 9], [30, 7, 1, 1, 4, 0, 4, 1, 0, 0, 72, 9], [33, 4, 1, 1, 8, 0, 4, 1, 0, 0, 45, 9], [73, 4, 1, 1, 5, 0, 4, 1, 0, 2246, 40, 9], [43, 5, 2, 1, 5, 5, 4, 0, 0, 0, 70, 9], [41, 4, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [45, 4, 1, 1, 5, 0, 4, 1, 0, 1977, 40, 9], [38, 4, 1, 1, 8, 0, 4, 1, 0, 0, 55, 9], [24, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [40, 7, 5, 1, 5, 1, 4, 0, 0, 0, 45, 9]], \"covered_false\": [[35, 4, 1, 1, 5, 5, 4, 0, 7298, 0, 8, 0], [44, 4, 1, 1, 8, 0, 4, 1, 0, 1902, 56, 9], [44, 4, 2, 1, 5, 0, 4, 1, 0, 0, 40, 9], [47, 7, 2, 1, 5, 0, 4, 1, 15024, 0, 50, 9], [46, 7, 2, 1, 5, 0, 4, 1, 7688, 0, 45, 9], [36, 4, 1, 1, 7, 0, 2, 1, 7298, 0, 36, 9], [28, 4, 1, 1, 5, 3, 4, 0, 0, 1564, 40, 9], [64, 5, 1, 1, 8, 0, 4, 1, 15024, 0, 55, 9], [48, 4, 1, 1, 8, 0, 4, 1, 0, 0, 55, 9], [51, 4, 1, 1, 6, 0, 4, 1, 7688, 0, 50, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[41, 2, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [45, 6, 1, 1, 5, 3, 4, 1, 0, 0, 45, 3], [41, 4, 5, 1, 1, 3, 4, 0, 0, 0, 40, 9], [22, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [49, 4, 2, 1, 6, 3, 4, 1, 0, 0, 60, 9], [28, 1, 1, 1, 5, 3, 2, 0, 0, 0, 40, 9], [41, 5, 1, 1, 4, 3, 4, 1, 0, 0, 40, 9], [38, 4, 1, 1, 2, 3, 4, 1, 4508, 0, 40, 9], [44, 4, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [32, 7, 1, 1, 1, 3, 4, 1, 0, 0, 37, 9]], \"covered_false\": [[37, 4, 1, 1, 5, 3, 4, 1, 10520, 0, 40, 9], [27, 4, 1, 1, 5, 3, 1, 1, 13550, 0, 40, 9], [39, 4, 1, 1, 8, 3, 4, 1, 99999, 0, 70, 9], [45, 4, 5, 1, 2, 3, 4, 1, 14344, 0, 48, 9], [59, 6, 1, 1, 6, 3, 4, 1, 15024, 0, 40, 9], [54, 7, 2, 1, 8, 3, 4, 1, 0, 1902, 50, 9], [44, 4, 5, 1, 5, 3, 4, 1, 7688, 0, 55, 9], [33, 4, 5, 1, 5, 3, 4, 1, 15024, 0, 44, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[56, 4, 1, 1, 5, 3, 4, 1, 0, 0, 50, 9], [38, 4, 1, 1, 6, 3, 4, 1, 0, 0, 60, 9], [32, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [63, 4, 1, 1, 5, 3, 4, 0, 0, 0, 36, 9], [36, 7, 5, 1, 5, 3, 4, 1, 0, 0, 30, 9], [49, 4, 5, 1, 8, 3, 4, 1, 0, 0, 40, 9], [43, 4, 1, 1, 6, 3, 4, 0, 0, 0, 50, 9], [48, 1, 5, 1, 8, 3, 4, 0, 0, 0, 50, 9], [61, 4, 2, 1, 5, 3, 4, 1, 0, 0, 40, 9], [40, 7, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9]], \"covered_false\": [], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\", \"Capital Gain <= 0.00\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
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

```python

```
