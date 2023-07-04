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

requests.post("http://localhost:8080/v2/models/sentiment-input-transform_1/infer", json=inference_request).json()
```

```json
{'model_name': 'sentiment',
 'id': 'd49609fb-d950-439c-b06f-3c455f03a2b4',
 'parameters': {'content_type': 'str'},
 'outputs': [{'name': 'output-1',
   'shape': [1, 1],
   'datatype': 'BYTES',
   'parameters': {'content_type': 'str'},
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

requests.post("http://localhost:8080/v2/models/sentiment-output-transform_1/infer", json=inference_request).json()
```

```json
{'model_name': 'sentiments',
 'id': '2a4eeab2-c19b-4455-8c2c-2ab68b429ec9',
 'parameters': {'content_type': 'np'},
 'outputs': [{'name': 'output-1',
   'shape': [2, 1],
   'datatype': 'INT64',
   'parameters': {'content_type': 'np'},
   'data': [1, 0]}]}

```

## Sentiment

```python
inference_request = {'model_name': 'sentiment',
 'id': 'd49609fb-d950-439c-b06f-3c455f03a2b4',
 'parameters': {'content_type': 'hf'},
 'inputs': [{'name': 'args',
   'shape': [1, 1],
   'datatype': 'BYTES',
   'parameters': {'content_type': 'hf'},
   'data': ['This is not amazing at all.']}]}

requests.post("http://localhost:8080/v2/models/sentiment_1/infer", json=inference_request).json()
```

```json
{'model_name': 'sentiment_1',
 'model_version': '1',
 'id': 'd49609fb-d950-439c-b06f-3c455f03a2b4',
 'parameters': {},
 'outputs': [{'name': 'output',
   'shape': [1, 1],
   'datatype': 'BYTES',
   'parameters': {'content_type': 'hg_jsonlist'},
   'data': ['{"label": "NEGATIVE", "score": 0.9997649788856506}']}]}

```

```python
inference_request = {'model_name': 'sentiment',
 'id': 'd49609fb-d950-439c-b06f-3c455f03a2b4',
 'inputs': [{'name': 'args',
   'shape': [1, 1],
   'datatype': 'BYTES',
   'parameters': {'content_type': 'str'},
   'data': ['This is not amazing at all.']}]}

requests.post("http://localhost:8080/v2/models/sentiment_1/infer", json=inference_request).json()
```

```json
{'model_name': 'sentiment_1',
 'model_version': '1',
 'id': 'd49609fb-d950-439c-b06f-3c455f03a2b4',
 'parameters': {},
 'outputs': [{'name': 'output',
   'shape': [1, 1],
   'datatype': 'BYTES',
   'parameters': {'content_type': 'hg_jsonlist'},
   'data': ['{"label": "NEGATIVE", "score": 0.9997649788856506}']}]}

```

```python
inference_request = {'model_name': 'sentiment',
 'id': 'd49609fb-d950-439c-b06f-3c455f03a2b4',
                     'parameters': {"content_type":"str"},
 'inputs': [{'name': 'args',
   'shape': [1, 1],
   'datatype': 'BYTES',
   'parameters': {'content_type': 'str'},
   'data': ['This is not amazing at all.']}]}

requests.post("http://localhost:8080/v2/models/sentiment_1/infer", json=inference_request).json()
```

```
---------------------------------------------------------------------------

```

```
JSONDecodeError                           Traceback (most recent call last)

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/requests/models.py:971, in Response.json(self, **kwargs)
    970 try:
--> 971     return complexjson.loads(self.text, **kwargs)
    972 except JSONDecodeError as e:
    973     # Catch JSON-related errors and raise as requests.JSONDecodeError
    974     # This aliases json.JSONDecodeError and simplejson.JSONDecodeError

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/json/__init__.py:346, in loads(s, cls, object_hook, parse_float, parse_int, parse_constant, object_pairs_hook, **kw)
    343 if (cls is None and object_hook is None and
    344         parse_int is None and parse_float is None and
    345         parse_constant is None and object_pairs_hook is None and not kw):
--> 346     return _default_decoder.decode(s)
    347 if cls is None:

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/json/decoder.py:337, in JSONDecoder.decode(self, s, _w)
    333 """Return the Python representation of ``s`` (a ``str`` instance
    334 containing a JSON document).
    335
    336 """
--> 337 obj, end = self.raw_decode(s, idx=_w(s, 0).end())
    338 end = _w(s, end).end()

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/json/decoder.py:355, in JSONDecoder.raw_decode(self, s, idx)
    354 except StopIteration as err:
--> 355     raise JSONDecodeError("Expecting value", s, err.value) from None
    356 return obj, end

```

```
JSONDecodeError: Expecting value: line 1 column 1 (char 0)

```

```
During handling of the above exception, another exception occurred:

```

```
JSONDecodeError                           Traceback (most recent call last)

```

```
Input In [36], in <cell line: 10>()
      1 inference_request = {'model_name': 'sentiment',
      2  'id': 'd49609fb-d950-439c-b06f-3c455f03a2b4',
      3                      'parameters': {"content_type":"str"},
   (...)
      7    'parameters': {'content_type': 'str'},
      8    'data': ['This is not amazing at all.']}]}
---> 10 requests.post("http://localhost:8080/v2/models/sentiment_1/infer", json=inference_request).json()

```

```
File ~/miniconda3/envs/scv2/lib/python3.9/site-packages/requests/models.py:975, in Response.json(self, **kwargs)
    971     return complexjson.loads(self.text, **kwargs)
    972 except JSONDecodeError as e:
    973     # Catch JSON-related errors and raise as requests.JSONDecodeError
    974     # This aliases json.JSONDecodeError and simplejson.JSONDecodeError
--> 975     raise RequestsJSONDecodeError(e.msg, e.doc, e.pos)

```

```
JSONDecodeError: Expecting value: line 1 column 1 (char 0)

```

```python
requests.get("http://localhost:8080/v2/models/sentiment_1").json()
```

```json
{'name': 'sentiment_1',
 'versions': [],
 'platform': '',
 'inputs': [{'name': 'args', 'datatype': 'BYTES', 'shape': [1]},
  {'name': 'args',
   'datatype': 'BYTES',
   'shape': [-1],
   'parameters': {'content_type': 'hf'}}],
 'outputs': [{'name': 'outputs',
   'datatype': 'BYTES',
   'shape': [-1],
   'parameters': {'content_type': 'hg_json'}}],
 'parameters': {}}

```

```python

```
