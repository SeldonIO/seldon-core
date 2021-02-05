BASE=../../..

## REST

run_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model.yaml --http_port 8000 --grpc_port 9500 -log_level DEBUG


run_dummy_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_local

curl_rest:
	curl -v localhost:8000/api/v0.1/predictions -d '{"data":{"ndarray":[[1.0,2.0]]}}' -H "Accept: application/json" -H "Content-Type: application/json"

curl_metrics:
	curl localhost:8000/metrics

grpc_test:
	cd ${BASE}/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' -plaintext -proto ./prediction.proto  0.0.0.0:9500 seldon.protos.Seldon/Predict
