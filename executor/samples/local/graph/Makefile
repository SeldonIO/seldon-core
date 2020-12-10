BASE=../../..

## REST

run_executor:
	${BASE}/executor --sdep seldon-model --namespace default --predictor example --file ./model.yaml --http_port 8000 --grpc_port 5000


run_model:
	export PREDICTIVE_UNIT_HTTP_SERVICE_PORT=9001 && export PREDICTIVE_UNIT_GRPC_SERVICE_PORT=9501 && export PREDICTIVE_UNIT_METRICS_SERVICE_PORT=6001 && seldon-core-microservice --service-type MODEL Model GRPC

run_transformer:
	export PREDICTIVE_UNIT_HTTP_SERVICE_PORT=9000 && export PREDICTIVE_UNIT_GRPC_SERVICE_PORT=9500 && seldon-core-microservice --service-type TRANSFORMER Transformer GRPC


curl_rest:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -H "Content-Type: application/json" -d '{"data":{"ndarray":[[1.0,2.0]]}}'

curl_rest_multipart:
	curl -v localhost:8000/api/v0.1/predictions -H "Accept: application/json" -F jsonData=@input.json


grpc_test:
	cd ${BASE}/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0]]}}' -plaintext -proto ./prediction.proto  0.0.0.0:5000 seldon.protos.Seldon/Predict



