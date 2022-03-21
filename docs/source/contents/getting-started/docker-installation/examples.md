# Local Examples

Run these examples from the `samples` folder.

## Model

Launch a simple ML model and send prediction requests.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
  namespace: seldon-mesh
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn
```

Load model

```bash
seldon model load -f ./models/sklearn-iris-gs.yaml -v=true
```

Wait until the model is ready.

```bash
seldon model status --model-name iris -w ModelAvailable
```

````{collapse} Expand to see output 
```json
{"iris":"ModelAvailable"}
```
````

Make a REST inference call to model.

```bash
seldon model infer --model-name iris \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output 
```json
{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "b778cfc6-6eb7-4e6d-820e-12082ba8cddd",
	"parameters": null,
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1
			],
			"datatype": "INT64",
			"parameters": null,
			"data": [
				2
			]
		}
	]
}
```
````

Unload the model.

```bash
seldon model unload --model-name iris
```

