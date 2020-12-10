BASE=../../..

## REST

run_executor:
	JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 ${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model.yaml --http_port 8000 --grpc_port 5000


run_dummy_model:
	cd ${BASE}/../examples/models/mean_classifier && make run_local 

curl_rest:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"ndarray":[[1.0,2.0]]}}' 

curl_status:
	curl -v localhost:8000/api/v0.1/status/classifier

curl_metadata:
	curl -v localhost:8000/api/v0.1/metadata/classifier

run_jaeger:
	docker run -p 6831:6831/udp -p 16686:16686 jaegertracing/all-in-one:latest


grpc_test:
	cd ${BASE}/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' -plaintext -proto ./prediction.proto  0.0.0.0:5000 seldon.protos.Seldon/Predict


