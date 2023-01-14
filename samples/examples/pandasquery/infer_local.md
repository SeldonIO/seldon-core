```bash
cat model.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: pandasquery
spec:
  storageUri: "/mnt/models/pandasquery"
  requirements:
  - mlserver
  - python
  parameters:
  - name: query
    value: "A == 1"

```

```bash
seldon model load -f model.yaml
```

```json
{}

```

```bash
seldon model status pandasquery -w ModelAvailable | jq .
```

```json
{}

```

```bash
seldon model infer pandasquery \
  '{"inputs": [{"name": "A", "shape": [1,3], "datatype": "FP32", "data": [4,1,3]}]}'
```

```json
{
	"model_name": "pandasquery_1",
	"model_version": "1",
	"id": "df4e1e04-f66f-41e8-a09e-fcf7a1a17f36",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "status",
			"shape": [
				1
			],
			"datatype": "BYTES",
			"parameters": null,
			"data": [
				"no rows satisfied A == 2"
			]
		}
	]
}

```

```bash
seldon model unload pandasquery
```

```json
{}

```

```python

```
