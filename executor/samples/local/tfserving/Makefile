BASE=../../..

## REST

run_rest_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_rest.yaml --http_port 8000 --protocol tensorflow

run_rest_executor_chain:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_rest_chain.yaml --http_port 8000 --protocol tensorflow


run_tensorflow_serving:
	docker run --name tfserver --rm -p 8501:8501 -p 8500:8500 -v "${PWD}/model:/models/half_plus_two" -e MODEL_NAME=half_plus_two tensorflow/serving

curl_rest:
	curl -d '{"instances": [1.0, 2.0, 5.0]}'     -X POST http://localhost:8000/v1/models/half_plus_two/:predict

curl_status:
	curl http://localhost:8000/v1/models/half_plus_two


## GRPC

run_grpc_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_grpc.yaml --http_port 8000 --grpc_port 5000 --transport grpc --protocol tensorflow

run_grpc_executor_chain:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model_grpc_chain.yaml --http_port 8000 --grpc_port 5000 --transport grpc --protocol tensorflow


grpc_test:
	cd ${BASE}/proto && grpcurl -d '{"model_spec":{"name":"half_plus_two"},"inputs":{"x":{"dtype": 1, "tensor_shape": {"dim":[{"size": 3}]}, "floatVal" : [1.0, 2.0, 3.0]}}}' -plaintext -proto ./prediction_service.proto  0.0.0.0:5000 tensorflow.serving.PredictionService/Predict

grpc_status:
	cd ${BASE}/proto && grpcurl -d '{"model_spec":{"name":"half_plus_two"}}' -plaintext -proto ./model_service.proto  0.0.0.0:5000 tensorflow.serving.ModelService/GetModelStatus


