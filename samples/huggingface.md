## Huggingface Examples

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
	"id": "95270c33-e2e3-4799-b243-06e834e730b8",
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
				"{\"generated_text\": \"Once upon a time in a galaxy far away?\\nI hope the experience will be filled with awe and excitement for our current hero's creation. I would love for you to give you an opportunity to enjoy the game and share some of the gameplay ideas\"}"
			]
		}
	]
}

```

```python
res = !seldon model infer text-gen --inference-mode grpc \
   '{"inputs":[{"name":"args","contents":{"bytes_contents":["T25jZSB1cG9uIGEgdGltZQo="]},"datatype":"BYTES","shape":[1]}]}'
```

```python
import json
import base64
r = json.loads(res[0])
base64.b64decode(r["outputs"][0]["contents"]["bytesContents"][0])
```

```
b'{"generated_text": "Once upon a time\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n\\n"}'

```

Unload the model

```bash
seldon model unload text-gen
```

### Custom Text Generation Model

```bash
cat ./models/hf-custom-text-gen.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: custom-text-gen
spec:
  storageUri: "gs://viktor-models/scv2/samples/mlserver_1.3.5/text-generation-huggingface"  # change bucket name
  requirements:
    - huggingface
  memory: 3Gi

```

Load the model

```bash
seldon model load -f ./models/hf-custom-text-gen.yaml
```

```json
{}

```

```bash
seldon model status custom-text-gen -w ModelAvailable | jq -M .
```

```json
{}

```

```bash
seldon model infer custom-text-gen \
  '{"inputs": [{"name": "args","shape": [1],"datatype": "BYTES","data": ["Once upon a time in a galaxy far away"]}]}'
```

```json
{
	"model_name": "custom-text-gen_1",
	"model_version": "1",
	"id": "726a4b73-25b6-4f6c-8473-02ab5eaa72fd",
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
				"{\"generated_text\": \"Once upon a time in a galaxy far away that the vast majority of our current civilization is of the same technological, racial, and culture, humanity may have found a way out of those limitations. This may be where the future is not at stake.\"}"
			]
		}
	]
}

```

```python
res = !seldon model infer custom-text-gen --inference-mode grpc \
   '{"inputs":[{"name":"args","contents":{"bytes_contents":["T25jZSB1cG9uIGEgdGltZQo="]},"datatype":"BYTES","shape":[1]}]}'
```

```python
import json
import base64
r = json.loads(res[0])
base64.b64decode(r["outputs"][0]["contents"]["bytesContents"][0])
```

```
b'{"generated_text": "Once upon a time\\n\\nIn the past, there were certain classes of adventurers and travelers, who sought to enter a single location and do one thing a little different than everybody else. Then there were those who were good folk, and those who had"}'

```

Unload the model

```bash
seldon model unload custom-text-gen
```
