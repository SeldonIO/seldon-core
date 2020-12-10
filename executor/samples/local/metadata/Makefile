BASE=../../..

EXECUTOR_PORT ?= 8000
LOCALHOST_PORT ?= 9000

GRAPH ?= deploy.yaml


run_executor:
	go run ${BASE}/cmd/executor/main.go --sdep seldon-model --namespace default --predictor example --file ./${GRAPH} --http_port 8000

run_executor_chain:
	go run ${BASE}/cmd/executor/main.go --sdep seldon-model --namespace default --predictor example --file ./chain.yaml --http_port 8000

run_executor_broken_chain:
	go run ${BASE}/cmd/executor/main.go --sdep seldon-model --namespace default --predictor example --file ./broken_chain.yaml --http_port 8000

run_executor_combiner:
	go run ${BASE}/cmd/executor/main.go --sdep seldon-model --namespace default --predictor example --file ./combiner.yaml --http_port 8000



run_mock_model:
	PREDICTIVE_UNIT_SERVICE_PORT=${LOCALHOST_PORT} seldon-core-microservice Model REST --service-type MODEL --log-level DEBUG

run_docker_compose:
	docker-compse up --build


request:
	curl -s curl -X POST -H 'Content-Type: application/json' \
		 -d '{"data": {"names": ["input"], "ndarray": ["data"]}}' \
		 http://localhost:${EXECUTOR_PORT}/api/v1.0/predictions | jq .


metadata:
	curl -s http://localhost:${EXECUTOR_PORT}/api/v1.0/metadata/mock-model | jq .


graph-metadata:
	curl -s http://localhost:${EXECUTOR_PORT}/api/v1.0/metadata | jq .


request-grpc:
	cd ${BASE}/proto && grpcurl \
		-d '{"data": {"names": ["input"], "ndarray": ["data"]}}' \
		-plaintext -proto ./prediction.proto  0.0.0.0:8000 seldon.protos.Seldon/Predict

metadata-grpc:
	cd ${BASE}/proto && grpcurl \
	-d '{"name": "mock-model"}' \
	-plaintext -proto ./prediction.proto  0.0.0.0:8000 seldon.protos.Seldon/ModelMetadata

graph-metadata-grpc:
	cd ${BASE}/proto && grpcurl \
	-plaintext -proto ./prediction.proto  0.0.0.0:8000 seldon.protos.Seldon/GraphMetadata
