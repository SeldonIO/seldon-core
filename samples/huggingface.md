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
	"id": "bcd24642-ade5-4b39-9c70-5d62bbfbe43c",
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
				"{\"generated_text\": \"Once upon a time in a galaxy far away, the Galaxy has made it a little easier to travel to and from your home planet through the galaxy's solar system. The planet's atmosphere is also a key asset to the Galaxy's galactic evolution! The\"}"
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
b'{"generated_text": "Once upon a time\\n\\n\\nI have no idea if he or she will fall prey to my whims\\nAnd\\n\\nAnd when\\n\\nOr\\n\\nAnd when I say\\nTo\\n\\nI\'ll have to\\nEven\\nOr\\nTo"}'

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
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.5/custom-text-generation-huggingface"
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
	"id": "ea5d4209-a5cf-4179-ba81-87b8bcfdf894",
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
				"{\"generated_text\": \"Once upon a time in a galaxy far away, scientists in the universe are told the universe is on a collision course with stars, and the result is the emergence of new universes. This has been possible for, say, the distant history of the Milky\"}"
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
b'{"generated_text": "Once upon a time\\n\\nand again\\n\\nwe took care not a minute of\\n\\nbeing alone in a\\n\\nworld that we had not seen\\n\\nfor\\n\\nten thousand years\\n\\nbefore it occurred to our\\n\\nmind that"}'

```

Unload the model

```bash
seldon model unload custom-text-gen
```
