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
	"id": "5412f8b0-f440-4e33-851c-ac3ff754b9c0",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "output",
			"shape": [
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "str",
				"headers": null
			},
			"data": [
				"[{\"generated_text\": \"Once upon a time in a galaxy far away, the Human Race was destroyed at the hands of an evil machine. But in a galaxy far away, the Human Race was destroyed at the hands of an evil machine. But in a galaxy far away,\"}]"
			]
		}
	]
}

```

```bash
seldon model infer text-gen --inference-mode grpc \
   '{"inputs":[{"name":"args","contents":{"bytes_contents":["T25jZSB1cG9uIGEgdGltZQo="]},"datatype":"BYTES","shape":[1]}]}'
```

```json
{"modelName":"text-gen_1", "modelVersion":"1", "outputs":[{"name":"output", "datatype":"BYTES", "shape":["1"], "parameters":{"content_type":{"stringParam":"str"}}, "contents":{"bytesContents":["W3siZ2VuZXJhdGVkX3RleHQiOiAiT25jZSB1cG9uIGEgdGltZVxuXG5Zb3UgaGF2ZSB0YWtlbiBhIGJpZyB0dW1ibGUgaW50byB5b3VyIGxpdmluZyByb29tIGFuZCBJIGNhbiBzZWUgdGhhdCB5b3UgYXJlIGxvb2tpbmcgYXQgaGltXG5Zb3Ugd2FudCB0byBzdG9wP1xuWW91IHNlZSBhbiBleWU/XG5Zb3Uga25vdyB3aGF0P1xuU29tZXRoaW5nJ3MgYmVlbiBoYXBwZW5pbmcgaW4geW91ciJ9XQ=="]}}]}

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
