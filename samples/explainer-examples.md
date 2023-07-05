## Anchor Tabular Explainer for SKLearn Income Model

```bash
cat ./models/income.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/income/classifier"
  requirements:
  - sklearn

```

```bash
seldon model load -f ./models/income.yaml
```

```json
{}

```

```bash
seldon model status income -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer income \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income_1",
	"model_version": "1",
	"id": "c65b8302-85af-4bac-aac5-91e3bedebee8",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"data": [
				0
			]
		}
	]
}

```

```bash
cat ./models/income-explainer.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-explainer
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.3.5/income/explainer"
  explainer:
    type: anchor_tabular
    modelRef: income

```

```bash
seldon model load -f ./models/income-explainer.yaml
```

```json
{}

```

```bash
seldon model status income-explainer -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer income-explainer \
  '{"inputs": [{"name": "predict", "shape": [1, 12], "datatype": "FP32", "data": [[47,4,1,1,1,3,4,1,0,0,40,9]]}]}'
```

```json
{
	"model_name": "income-explainer_1",
	"model_version": "1",
	"id": "a22c3785-ff3b-4504-9b3c-199aa48a62d6",
	"parameters": {},
	"outputs": [
		{
			"name": "explanation",
			"shape": [
				1,
				1
			],
			"datatype": "BYTES",
			"parameters": {
				"content_type": "str"
			},
			"data": [
				"{\"meta\": {\"name\": \"AnchorTabular\", \"type\": [\"blackbox\"], \"explanations\": [\"local\"], \"params\": {\"seed\": 1, \"disc_perc\": [25, 50, 75], \"threshold\": 0.95, \"delta\": 0.1, \"tau\": 0.15, \"batch_size\": 100, \"coverage_samples\": 10000, \"beam_size\": 1, \"stop_on_first\": false, \"max_anchor_size\": null, \"min_samples_start\": 100, \"n_covered_ex\": 10, \"binary_cache_size\": 10000, \"cache_margin\": 1000, \"verbose\": false, \"verbose_every\": 1, \"kwargs\": {}}, \"version\": \"0.9.0\"}, \"data\": {\"anchor\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\"], \"precision\": 0.9518716577540107, \"coverage\": 0.07165109034267912, \"raw\": {\"feature\": [3, 5], \"mean\": [0.7959381044487428, 0.9518716577540107], \"precision\": [0.7959381044487428, 0.9518716577540107], \"coverage\": [0.3037383177570093, 0.07165109034267912], \"examples\": [{\"covered_true\": [[52, 5, 5, 1, 8, 1, 2, 0, 0, 0, 50, 9], [49, 4, 1, 1, 4, 4, 1, 0, 0, 0, 40, 1], [23, 4, 1, 1, 6, 1, 4, 1, 0, 0, 40, 9], [55, 2, 1, 1, 5, 1, 4, 0, 0, 0, 48, 9], [22, 4, 1, 1, 2, 3, 4, 0, 0, 0, 15, 9], [51, 4, 2, 1, 5, 0, 1, 1, 0, 0, 99, 4], [40, 4, 1, 1, 5, 1, 4, 0, 0, 0, 40, 9], [40, 6, 1, 1, 2, 0, 4, 1, 0, 0, 50, 9], [50, 5, 5, 1, 6, 0, 4, 1, 0, 0, 55, 9], [41, 4, 1, 1, 6, 0, 4, 1, 0, 0, 40, 9]], \"covered_false\": [[42, 4, 1, 1, 8, 0, 4, 1, 0, 2415, 60, 9], [48, 6, 2, 1, 5, 4, 4, 0, 0, 0, 60, 9], [37, 4, 1, 1, 5, 0, 4, 1, 0, 0, 45, 9], [57, 4, 5, 1, 8, 0, 4, 1, 0, 0, 50, 9], [63, 7, 2, 1, 8, 0, 4, 1, 0, 1902, 50, 9], [51, 4, 5, 1, 8, 0, 4, 1, 0, 1887, 47, 9], [51, 2, 2, 1, 8, 1, 4, 0, 0, 0, 45, 9], [68, 7, 5, 1, 5, 0, 4, 1, 0, 2377, 42, 0], [45, 4, 1, 1, 8, 0, 4, 1, 15024, 0, 40, 9], [45, 4, 1, 1, 8, 0, 4, 1, 0, 1977, 60, 9]], \"uncovered_true\": [], \"uncovered_false\": []}, {\"covered_true\": [[44, 6, 5, 1, 8, 3, 4, 0, 0, 1902, 60, 9], [58, 7, 2, 1, 5, 3, 1, 1, 4064, 0, 40, 1], [50, 7, 1, 1, 1, 3, 2, 0, 0, 0, 37, 9], [34, 4, 2, 1, 5, 3, 4, 1, 0, 0, 45, 9], [45, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [33, 7, 5, 1, 5, 3, 1, 1, 0, 0, 30, 6], [61, 7, 2, 1, 5, 3, 4, 1, 0, 0, 40, 0], [35, 4, 5, 1, 1, 3, 4, 1, 0, 0, 40, 9], [71, 2, 1, 1, 5, 3, 4, 0, 0, 0, 6, 9], [44, 4, 1, 1, 8, 3, 2, 1, 0, 0, 35, 9]], \"covered_false\": [[30, 4, 5, 1, 5, 3, 4, 1, 10520, 0, 40, 9], [54, 7, 2, 1, 8, 3, 4, 1, 0, 1902, 50, 9], [66, 6, 2, 1, 6, 3, 4, 1, 0, 2377, 25, 9], [35, 4, 2, 1, 5, 3, 4, 1, 7298, 0, 40, 9], [44, 4, 1, 1, 8, 3, 4, 1, 7298, 0, 48, 9], [31, 4, 1, 1, 8, 3, 4, 0, 13550, 0, 50, 9], [35, 4, 1, 1, 8, 3, 4, 1, 8614, 0, 45, 9]], \"uncovered_true\": [], \"uncovered_false\": []}], \"all_precision\": 0, \"num_preds\": 1000000, \"success\": true, \"names\": [\"Marital Status = Never-Married\", \"Relationship = Own-child\"], \"prediction\": [0], \"instance\": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], \"instances\": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}"
			]
		}
	]
}

```

```bash
seldon model unload income-explainer
```

```json
{}

```

```bash
seldon model unload income
```

```json
{}

```

## Anchor Text Explainer for SKLearn Movies Sentiment Model

```bash
cat ./models/moviesentiment.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: sentiment
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.5/moviesentiment-sklearn"
  requirements:
  - sklearn

```

```bash
seldon model load -f ./models/moviesentiment.yaml
```

```json
{}

```

```bash
seldon model status sentiment -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer sentiment \
  '{"parameters": {"content_type": "str"}, "inputs": [{"name": "foo", "data": ["I am good"], "datatype": "BYTES","shape": [1]}]}'
```

```json
{
	"model_name": "sentiment_2",
	"model_version": "1",
	"id": "f5c07363-7e9d-4f09-aa30-228c81fdf4a4",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				0
			]
		}
	]
}

```

```bash
cat ./models/moviesentiment-explainer.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: sentiment-explainer
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.3.5/moviesentiment-sklearn-explainer"
  explainer:
    type: anchor_text
    modelRef: sentiment

```

```bash
seldon model load -f ./models/moviesentiment-explainer.yaml
```

```json
{}

```

```bash
seldon model status sentiment-explainer -w ModelAvailable
```

```json
{}

```

```bash
seldon model infer sentiment-explainer \
  '{"parameters": {"content_type": "str"}, "inputs": [{"name": "foo", "data": ["I am good"], "datatype": "BYTES","shape": [1]}]}'
```

```
Error: V2 server error: 500 Traceback (most recent call last):
  File "/opt/conda/lib/python3.8/site-packages/starlette/middleware/errors.py", line 162, in __call__
    await self.app(scope, receive, _send)
  File "/opt/conda/lib/python3.8/site-packages/starlette_exporter/middleware.py", line 307, in __call__
    await self.app(scope, receive, wrapped_send)
  File "/opt/conda/lib/python3.8/site-packages/starlette/middleware/gzip.py", line 24, in __call__
    await responder(scope, receive, send)
  File "/opt/conda/lib/python3.8/site-packages/starlette/middleware/gzip.py", line 44, in __call__
    await self.app(scope, receive, self.send_with_gzip)
  File "/opt/conda/lib/python3.8/site-packages/starlette/middleware/exceptions.py", line 79, in __call__
    raise exc
  File "/opt/conda/lib/python3.8/site-packages/starlette/middleware/exceptions.py", line 68, in __call__
    await self.app(scope, receive, sender)
  File "/opt/conda/lib/python3.8/site-packages/fastapi/middleware/asyncexitstack.py", line 21, in __call__
    raise e
  File "/opt/conda/lib/python3.8/site-packages/fastapi/middleware/asyncexitstack.py", line 18, in __call__
    await self.app(scope, receive, send)
  File "/opt/conda/lib/python3.8/site-packages/starlette/routing.py", line 706, in __call__
    await route.handle(scope, receive, send)
  File "/opt/conda/lib/python3.8/site-packages/starlette/routing.py", line 276, in handle
    await self.app(scope, receive, send)
  File "/opt/conda/lib/python3.8/site-packages/starlette/routing.py", line 66, in app
    response = await func(request)
  File "/opt/conda/lib/python3.8/site-packages/mlserver/rest/app.py", line 42, in custom_route_handler
    return await original_route_handler(request)
  File "/opt/conda/lib/python3.8/site-packages/fastapi/routing.py", line 237, in app
    raw_response = await run_endpoint_function(
  File "/opt/conda/lib/python3.8/site-packages/fastapi/routing.py", line 163, in run_endpoint_function
    return await dependant.call(**values)
  File "/opt/conda/lib/python3.8/site-packages/mlserver/rest/endpoints.py", line 99, in infer
    inference_response = await self._data_plane.infer(
  File "/opt/conda/lib/python3.8/site-packages/mlserver/handlers/dataplane.py", line 103, in infer
    prediction = await model.predict(payload)
  File "/opt/conda/lib/python3.8/site-packages/mlserver_alibi_explain/runtime.py", line 86, in predict
    output_data = await self._async_explain_impl(input_data, payload.parameters)
  File "/opt/conda/lib/python3.8/site-packages/mlserver_alibi_explain/runtime.py", line 119, in _async_explain_impl
    explanation = await loop.run_in_executor(self._executor, explain_call)
  File "/opt/conda/lib/python3.8/concurrent/futures/thread.py", line 57, in run
    result = self.fn(*self.args, **self.kwargs)
  File "/opt/conda/lib/python3.8/site-packages/mlserver_alibi_explain/explainers/black_box_runtime.py", line 62, in _explain_impl
    input_data = input_data[0]
KeyError: 0

```

```bash
seldon model unload sentiment-explainer
```

```json
{}

```

```bash
seldon model unload sentiment
```

```json
{}

```

