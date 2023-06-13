## Rolling Update Examples

### SKLearn Iris Model

We use a simple sklearn iris classification model and do a rolling update

```bash
seldon model load -f ./models/iris-v1.yaml
seldon model status iris -w ModelAvailable
seldon model infer iris -i 1 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```json
{}
{}
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "e36b1d20-ddbb-4bbc-9649-290a8bb14ac4",
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
				2
			]
		}
	]
}

```

```bash
seldon model load -f ./models/iris-v2.yaml
seldon model infer iris --seconds 5 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
{}
Success: map[:iris_1::216 :iris_2::984]

```

```bash
seldon model unload iris
```

```json
{}

```

```python

```
