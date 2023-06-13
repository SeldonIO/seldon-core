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
	"id": "bde3e68b-2710-4763-828c-21e58416b45c",
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
				"content_type": "str"
			},
			"data": [
				"[{\"generated_text\": \"Once upon a time in a galaxy far away, you can see the galaxy and the universe together.\\n\\n\\n\\nIn the same space as Earth, there are two locations (including Earth) or regions. These regions are located on both of the\"}]"
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
b'[{"generated_text": "Once upon a time\\n\\n\\nThe Great Depression was a devastating economic collapse. The depression was not only due to soaring interest rates, but to an epidemic of illegal foreign trade. It was the worst recession of all time because of the government\'s actions"}]'

```

Unload the model

```bash
seldon model unload text-gen
```

```json
{}

```

```python

```
