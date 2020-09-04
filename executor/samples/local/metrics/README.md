# Test Executor with Prometheus Metrics

## REST

Run the following commands in different terminals.

Start the executor locally.
```bash
make run_executor
```

Start a dummy REST model locally.
```bash
make run_dummy_rest_model
```

Send a request
```bash
make curl_rest
```

Check the metrics endpoint:

```
make curl_metrics
```

You should see metrics including:

```
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="0.005"} 0
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="0.01"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="0.025"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="0.05"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="0.1"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="0.25"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="0.5"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="1"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="2.5"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="5"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="10"} 1
seldon_api_executor_client_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict",le="+Inf"} 1
seldon_api_executor_client_requests_seconds_sum{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict"} 0.006920656
seldon_api_executor_client_requests_seconds_count{code="200",deployment_name="seldon-model",method="post",model_image="seldonio/mock_classifier_rest",model_name="classifier",model_version="1.3",predictor_name="example",predictor_version="",service="/predict"} 1
# HELP seldon_api_executor_server_requests_seconds A histogram of latencies for executor server
# TYPE seldon_api_executor_server_requests_seconds histogram
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="0.005"} 0
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="0.01"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="0.025"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="0.05"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="0.1"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="0.25"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="0.5"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="1"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="2.5"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="5"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="10"} 1
seldon_api_executor_server_requests_seconds_bucket{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions",le="+Inf"} 1
seldon_api_executor_server_requests_seconds_sum{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions"} 0.007476718
seldon_api_executor_server_requests_seconds_count{code="200",deployment_name="seldon-model",method="post",predictor_name="example",predictor_version="",service="/api/v0.1/predictions"} 1
```


## gRPC

Run the following commands in different terminals.


Start the executor locally.
```bash
make run_executor
```

Start a dummy REST model locally.
```bash
make run_dummy_grpc_model
```

Send a request
```bash
make grpc_test
```

Check the metrics endpoint:

```
make curl_metrics
```

You should see metrics including:

```
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="0.005"} 0
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="0.01"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="0.025"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="0.05"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="0.1"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="0.25"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="0.5"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="1"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="2.5"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="5"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="10"} 1
seldon_api_executor_client_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier",le="+Inf"} 1
seldon_api_executor_client_requests_seconds_sum{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier"} 0.005590603
seldon_api_executor_client_requests_seconds_count{code="OK",deployment_name="seldon-model",method="unary",model_image="1.3",model_name="seldonio/mock_classifier_rest",model_version="/seldon.protos.Model/Predict",predictor_name="example",predictor_version="",service="classifier"} 1
# HELP seldon_api_executor_server_requests_seconds A histogram of latencies for executor server
# TYPE seldon_api_executor_server_requests_seconds histogram
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="0.005"} 0
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="0.01"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="0.025"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="0.05"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="0.1"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="0.25"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="0.5"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="1"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="2.5"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="5"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="10"} 1
seldon_api_executor_server_requests_seconds_bucket{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict",le="+Inf"} 1
seldon_api_executor_server_requests_seconds_sum{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict"} 0.005860215
seldon_api_executor_server_requests_seconds_count{code="OK",deployment_name="seldon-model",method="unary",predictor_name="example",predictor_version="",service="/seldon.protos.Seldon/Predict"} 1

```