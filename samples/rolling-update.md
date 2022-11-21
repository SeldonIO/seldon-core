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
    	"id": "20aee4f5-fbc0-4406-9b2f-04e729f6bc78",
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
    Success: map[:iris_1::428 :iris_2::1958]



```python
!seldon model unload iris
```

    {}



```python

```
