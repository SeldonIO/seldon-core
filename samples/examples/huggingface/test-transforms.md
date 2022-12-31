# Test transforms

We need transforms to change the input to the sentiment model and the output from the sentiment model to allow for easier integration with the Alibi Explain explainer. The input to the explainer will come from the output of the speech recognition model which returns a Dict so we extract the text and turn into a simple list. Similarly, the sentiment model returns a Dict so we convert to a simple binary classifer result.

```python
import requests
```

## Sentiment Input Transform

run `mlserver start .` from the `sentiment-input-transform` folder.

This code turns the Dict into a simple list of strings.

```python
inference_request = {
    "inputs": [
        {
          "name": "predict",
          "shape": [1],
          "datatype": "BYTES",
          "data": ['{"text": "This is not amazing at all."}'],
        }
    ]
}

requests.post("http://localhost:8080/v2/models/sentiment-input-transform/infer", json=inference_request).json()
```

```json
{'model_name': 'sentiment',
 'model_version': None,
 'id': '52258223-b1bf-425f-88dd-d5edfe513b67',
 'parameters': {'content_type': None, 'headers': None},
 'outputs': [{'name': 'output-1',
   'shape': [1],
   'datatype': 'BYTES',
   'parameters': {'content_type': 'str', 'headers': None},
   'data': ['This is not amazing at all.']}]}

```

## Sentiment Output Transform

run `mlserver start .` from the `sentiment-output-transform` folder.

This code turns the Dict result from the sentiment model and returns a simple 1 or 0 classifier prediction.

```python
inference_request = {
    "inputs": [
        {
          "name": "predict",
          "shape": [1],
          "datatype": "BYTES",
          "data": ['{"label": "POSITIVE", "score":"0.99"}','{"label": "NEGATIVE", "score":"0.99"}'],
        }
    ]
}

requests.post("http://localhost:8080/v2/models/sentiment-output-transform/infer", json=inference_request).json()
```

```json
{'model_name': 'sentiments',
 'model_version': None,
 'id': 'a6937c55-05e7-4715-afb6-8b9d87b0fe1d',
 'parameters': {'content_type': None, 'headers': None},
 'outputs': [{'name': 'output-1',
   'shape': [2],
   'datatype': 'INT64',
   'parameters': None,
   'data': [1, 0]}]}

```

```python

```
