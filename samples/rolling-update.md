## Rolling Update Examples


### SKLearn Iris Model

We use a simple sklearn iris classification model and do a rolling update


```python
!seldon model load -f ./models/iris-v1.yaml
!seldon model status iris -w ModelAvailable 
!seldon model infer iris -i 1 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

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



```python
!seldon model load -f ./models/iris-v2.yaml
!seldon model infer iris --seconds 5 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    {}
    Success: map[:iris_1::362 :iris_2::1782]



```python
!seldon model unload iris
```

    {}



```python

```
