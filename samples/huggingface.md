## Huggingface Examples


### Text Generation Model



```python
!cat ./models/hf-text-gen.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: text-gen
    spec:
      storageUri: "gs://seldon-models/mlserver/huggingface/text-generation"
      requirements:
      - huggingface


Load the model


```python
!seldon model load -f ./models/hf-text-gen.yaml
```

    {}



```python
!seldon model status text-gen -w ModelAvailable | jq -M .
```

    {}



```python
!seldon model infer text-gen \
  '{"inputs": [{"name": "args","shape": [1],"datatype": "BYTES","data": ["Once upon a time in a galaxy far away"]}]}' 
```

    {
    	"model_name": "text-gen_1",
    	"model_version": "1",
    	"id": "951a5db1-2511-4304-b6a1-1145c5d7ba1a",
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
    				"{\"generated_text\": \"Once upon a time in a galaxy far away from the Milky Way, it looks very different. Scientists have discovered that star clusters are at least 1.5 times smaller than the sun and contain the biggest known mass of all the solar masses in the universe\"}"
    			]
    		}
    	]
    }



```python
!seldon model infer text-gen --inference-mode grpc \
   '{"inputs":[{"name":"args","contents":{"bytes_contents":["T25jZSB1cG9uIGEgdGltZQo="]},"datatype":"BYTES","shape":[1]}]}' 
```

    {"modelName":"text-gen_1", "modelVersion":"1", "outputs":[{"name":"output", "datatype":"BYTES", "shape":["1"], "parameters":{"content_type":{"stringParam":"str"}}, "contents":{"bytesContents":["eyJnZW5lcmF0ZWRfdGV4dCI6ICJPbmNlIHVwb24gYSB0aW1lXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG5cblxuXG4ifQ=="]}}]}


Unload the model


```python
!seldon model unload text-gen
```

    {}



```python

```
