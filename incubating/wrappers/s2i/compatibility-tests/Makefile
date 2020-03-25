KIND_NAME ?= seldon
AMBASSADOR_PORT ?= 8003


SELDON_CORE_DIR ?= ../../../..



images-rest:
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.19-SNAPSHOT compat-rest-model:0.19-SNAPSHOT
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.18 compat-rest-model:0.18
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.17-SNAPSHOT compat-rest-model:0.17
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.16 compat-rest-model:0.16
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.15 compat-rest-model:0.15
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.14 compat-rest-model:0.14
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.13 compat-rest-model:0.13
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.12 compat-rest-model:0.12
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.11 compat-rest-model:0.11
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python3:0.10 compat-rest-model:0.10

images-grpc:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.19-SNAPSHOT compat-grpc-model:0.19-SNAPSHOT
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.18 compat-grpc-model:0.18
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.17-SNAPSHOT compat-grpc-model:0.17
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.16 compat-grpc-model:0.16
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.15 compat-grpc-model:0.15
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.14 compat-grpc-model:0.14
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.13 compat-grpc-model:0.13
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.12 compat-grpc-model:0.12
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.11 compat-grpc-model:0.11
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python3:0.10 compat-grpc-model:0.10


kind-load-images-rest: images-rest
	kind load docker-image compat-rest-model:0.19-SNAPSHOT --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.18 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.17 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.16 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.15 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.14 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.13 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.12 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.11 --name ${KIND_NAME}
	kind load docker-image compat-rest-model:0.10 --name ${KIND_NAME}


kind-load-images-grpc: images-grpc
	kind load docker-image compat-grpc-model:0.19-SNAPSHOT --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.18 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.17 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.16 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.15 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.14 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.13 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.12 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.11 --name ${KIND_NAME}
	kind load docker-image compat-grpc-model:0.10 --name ${KIND_NAME}



images-java:
	s2i build -E environment_rest ${SELDON_CORE_DIR}/incubating/wrappers/s2i/java/test/model-template-app seldonio/seldon-core-s2i-java-build:0.1 compat-rest-java:0.1 --runtime-image seldonio/seldon-core-s2i-java-runtime:0.1
	s2i build -E environment_rest ${SELDON_CORE_DIR}/incubating/wrappers/s2i/java/test/model-template-app seldonio/seldon-core-s2i-java-build:0.2 compat-rest-java:0.2 --runtime-image seldonio/seldon-core-s2i-java-runtime:0.2

	s2i build -E environment_grpc ${SELDON_CORE_DIR}/incubating/wrappers/s2i/java/test/model-template-app seldonio/seldon-core-s2i-java-build:0.1 compat-grpc-java:0.1 --runtime-image seldonio/seldon-core-s2i-java-runtime:0.1
	s2i build -E environment_grpc ${SELDON_CORE_DIR}/incubating/wrappers/s2i/java/test/model-template-app seldonio/seldon-core-s2i-java-build:0.2 compat-grpc-java:0.2 --runtime-image seldonio/seldon-core-s2i-java-runtime:0.2

kind-load-images-java: images-java
	kind load docker-image compat-rest-java:0.1 --name ${KIND_NAME}
	kind load docker-image compat-rest-java:0.2 --name ${KIND_NAME}
	kind load docker-image compat-grpc-java:0.1 --name ${KIND_NAME}
	kind load docker-image compat-grpc-java:0.2 --name ${KIND_NAME}


deploy-rest:
	kubectl apply -f deploy-rest.yaml

deploy-grpc:
	kubectl apply -f deploy-grpc.yaml

deploy-java-rest:
	kubectl apply -f deploy-java-rest.yaml

deploy-java-grpc:
	kubectl apply -f deploy-java-grpc.yaml



remove-rest:
	kubectl delete -f deploy-rest.yaml

remove-grpc:
	kubectl delete -f deploy-grpc.yaml

remove-java-rest:
	kubectl delete -f deploy-java-rest.yaml

remove-java-grpc:
	kubectl delete -f deploy-java-grpc.yaml



request-rest:
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-019-snapshot-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-019-snapshot-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-018-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-018-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-017-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-017-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-016-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-016-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-015-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-015-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-014-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-014-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-013-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-013-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-012-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-012-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-011-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-011-engine/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-010-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"meta": {"tags": {"foo":"bar", "a": -1}}, "data": {"names": ["input"], "ndarray": ["data"]}}' \
		http://localhost:${AMBASSADOR_PORT}/seldon/seldon/compat-rest-model-010-engine/api/v1.0/predictions \
		| jq .


request-grpc:
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-019-snapshot-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto  0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-019-snapshot-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-018-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-018-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-017-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-017-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-016-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-016-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-015-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-015-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-014-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-014-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-013-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-013-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-012-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-012-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-011-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-011-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-010-executor -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl -d '{"data": {"names": ["get"], "ndarray": ["data"]}}' -rpc-header seldon:compat-grpc-model-010-engine -rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 seldon.protos.Seldon/Predict | jq .

request-java-rest:
	curl -s -X POST \
		-d 'json={"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		localhost:8003/seldon/seldon/compat-rest-java-02-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		localhost:8003/seldon/seldon/compat-rest-java-02-engine/api/v1.0/predictions \
		| jq .

	curl -s -X POST \
		-d 'json={"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		localhost:8003/seldon/seldon/compat-rest-java-01-executor/api/v1.0/predictions \
		| jq .
	curl -s -X POST -H 'Content-Type: application/json' \
		-d '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		localhost:8003/seldon/seldon/compat-rest-java-01-engine/api/v1.0/predictions \
		| jq .

request-java-grpc:
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl \
		-d '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		-rpc-header seldon:compat-grpc-java-02-executor \
		-rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 \
		seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl \
		-d '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		-rpc-header seldon:compat-grpc-java-02-engine \
		-rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 \
		seldon.protos.Seldon/Predict | jq .

	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl \
		-d '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		-rpc-header seldon:compat-grpc-java-01-executor \
		-rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 \
		seldon.protos.Seldon/Predict | jq .
	cd ${SELDON_CORE_DIR}/executor/proto && grpcurl \
		-d '{"data":{"names":["a","b"],"ndarray":[[1.0,2.0]]}}' \
		-rpc-header seldon:compat-grpc-java-01-engine \
		-rpc-header namespace:seldon -plaintext -proto ./prediction.proto 0.0.0.0:8003 \
		seldon.protos.Seldon/Predict | jq .



grep-test:
	cat deploy-rest.yaml | grep "image:"
	cat deploy-rest.yaml | grep "name: compat"
	cat deploy-rest.yaml | grep "seldon.io/executor"
	echo "\n\n"
	cat deploy-grpc.yaml | grep "image:"
	cat deploy-grpc.yaml | grep "name: compat"
	cat deploy-grpc.yaml | grep "seldon.io/executor"
	echo "\n\n"
	cat deploy-java-rest.yaml | grep "image:"
	cat deploy-java-rest.yaml | grep "name: compat"
	cat deploy-java-rest.yaml | grep "seldon.io/executor"
	echo "\n\n"
	cat deploy-java-grpc.yaml | grep "image:"
	cat deploy-java-grpc.yaml | grep "name: compat"
	cat deploy-java-grpc.yaml | grep "seldon.io/executor"
