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
    	"id": "d17c062e-6c19-43a6-812a-43aa46c2b109",
    	"parameters": {
    		"content_type": null,
    		"headers": null
    	},
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

```bash
seldon model load -f ./models/iris-v2.yaml
seldon model infer iris --seconds 5 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
```json
    {}
    Success: map[:iris_1::362 :iris_2::1782]
```

```bash
seldon model unload iris
```
```json
    {}
```

```python

```
