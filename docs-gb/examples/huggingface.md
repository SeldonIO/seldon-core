# Huggingface Examples

### Text Generation Model

```bash
cat ./models/hf-text-gen.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: text-gen
spec:
  storageUri: "gs://seldon-models/mlserver/huggingface/text-generation"
  requirements:
  - huggingface

```

Load the model

```bash
seldon model load -f ./models/hf-text-gen.yaml
```

```json
{}

```

```bash
seldon model status text-gen -w ModelAvailable | jq -M .
```

```json
{}

```

```bash
seldon model infer text-gen \
  '{"inputs": [{"name": "args","shape": [1],"datatype": "BYTES","data": ["Once upon a time in a galaxy far away"]}]}'
```

```json
{
	"model_name": "text-gen_1",
	"model_version": "1",
	"id": "121ff5f4-1d4a-46d0-9a5e-4cd3b11040df",
	"parameters": {},
	"outputs": [
		{
			"name": "output",
			"shape": [
				1,
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "hg_jsonlist"
			},
			"data": [
				"{\"generated_text\": \"Once upon a time in a galaxy far away, the planet is full of strange little creatures. A very strange combination of creatures in that universe, that is. A strange combination of creatures in that universe, that is. A kind of creature that is\"}"
			]
		}
	]
}

```

```python
res = !seldon model infer text-gen --inference-mode grpc \
   '{"inputs":[{"name":"args","contents":{"bytes_contents":["T25jZSB1cG9uIGEgdGltZSBpbiBhIGdhbGF4eSBmYXIgYXdheQo="]},"datatype":"BYTES","shape":[1]}]}'
```

```python
import json
import base64
r = json.loads(res[0])
base64.b64decode(r["outputs"][0]["contents"]["bytesContents"][0])
```

```
b'{"generated_text": "Once upon a time in a galaxy far away\\n\\nThe Universe is a big and massive place. How can you feel any of this? Your body doesn\'t make sense if the Universe is in full swing \\u2014 you don\'t have to remember whether the"}'

```

Unload the model

```bash
seldon model unload text-gen
```

### Custom Text Generation Model

```bash
cat ./models/hf-text-gen-custom-tiny-stories.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: custom-tiny-stories-text-gen
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.5/huggingface-text-gen-custom-tiny-stories"
  requirements:
    - huggingface

```

Load the model

```bash
seldon model load -f ./models/hf-text-gen-custom-tiny-stories.yaml
```

```json
{}

```

```bash
seldon model status custom-tiny-stories-text-gen -w ModelAvailable | jq -M .
```

```json
{}

```

```bash
seldon model infer custom-tiny-stories-text-gen \
  '{"inputs": [{"name": "args","shape": [1],"datatype": "BYTES","data": ["Once upon a time in a galaxy far away"]}]}'
```

```json
{
	"model_name": "custom-tiny-stories-text-gen_1",
	"model_version": "1",
	"id": "d0fce59c-76e2-4f81-9711-1c93d08bcbf9",
	"parameters": {},
	"outputs": [
		{
			"name": "output",
			"shape": [
				1,
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "hg_jsonlist"
			},
			"data": [
				"{\"generated_text\": \"Once upon a time in a galaxy far away. It was a very special place to live.\\n\"}"
			]
		}
	]
}

```

```python
res = !seldon model infer custom-tiny-stories-text-gen --inference-mode grpc \
   '{"inputs":[{"name":"args","contents":{"bytes_contents":["T25jZSB1cG9uIGEgdGltZSBpbiBhIGdhbGF4eSBmYXIgYXdheQo="]},"datatype":"BYTES","shape":[1]}]}'
```

```python
import json
import base64
r = json.loads(res[0])
base64.b64decode(r["outputs"][0]["contents"]["bytesContents"][0])
```

```
b'{"generated_text": "Once upon a time in a galaxy far away\\nOne night, a little girl named Lily went to"}'

```

Unload the model

```bash
seldon model unload custom-tiny-stories-text-gen
```

````
As a next step, why not try running a larger-scale model? You can find a definition for one in ./models/hf-text-gen-custom-gpt2.yaml. However, you may need to request and allocate more memory!
````
