BASE=../../..

triton-inference-server:
	git clone https://github.com/NVIDIA/triton-inference-server

triton-inference-server/docs/examples/model_repository/simple: triton-inference-server
	cd triton-inference-server/docs/examples && ./fetch_models.sh

## REST

run_rest_executor:
	${BASE}/executor --sdep triton --namespace default --predictor simple --file ./model_rest.yaml --port 8000 --protocol kfserving


run_triton_simple_rest:
	docker run --rm --shm-size=1g --ulimit memlock=-1 --ulimit stack=67108864 -p9000:9000 -p8001:8001 -p8002:8002 -v ${PWD}/triton-inference-server/docs/examples/model_repository:/models nvcr.io/nvidia/tritonserver:20.08-py3 /opt/tritonserver/bin/tritonserver --model-repository=/models --http-port=9000

curl_rest_triton:
	curl -v -d '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'    -X POST http://0.0.0.0:9000/v2/models/simple/infer    -H "Content-Type: application/json"

curl_rest:
	curl -v -d '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'    -X POST http://0.0.0.0:8000/v2/models/simple/infer    -H "Content-Type: application/json"


curl_status:
	curl -v http://localhost:8000/v2/models/simple/ready

curl_metadata:
	curl http://localhost:8000/v2/models/simple


## GRPC

run_grpc_executor:
	${BASE}/executor --sdep triton --namespace default --predictor simple --file ./model_rest.yaml --port 8000 --protocol kfserving

run_triton_simple_grpc:
	docker run --rm --shm-size=1g --ulimit memlock=-1 --ulimit stack=67108864 -p9000:9000 -p8001:8001 -p8002:8002 -v ${PWD}/triton-inference-server/docs/examples/model_repository:/models nvcr.io/nvidia/tritonserver:20.08-py3 /opt/tritonserver/bin/tritonserver --model-repository=/models --grpc-port=9000


grpc_test:
	cd ${BASE}/api/grpc/kfserving/inference && grpcurl -d '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' -plaintext -proto ./grpc_service.proto  0.0.0.0:8000 inference.GRPCInferenceService/ModelInfer

grpc_status:
	cd ${BASE}/api/grpc/kfserving/inference && grpcurl -d '{"name":"simple"}' -plaintext -proto ./grpc_service.proto  0.0.0.0:8000 inference.GRPCInferenceService/ModelReady

grpc_status_triton:
	cd ${BASE}/api/grpc/kfserving/inference && grpcurl -d '{"name":"simple"}' -plaintext -proto ./grpc_service.proto  0.0.0.0:8000 inference.GRPCInferenceService/ModelReady


grpc_metadata:
	cd ${BASE}/api/grpc/kfserving/inference && grpcurl -d '{"name":"simple"}' -plaintext -proto ./grpc_service.proto  0.0.0.0:8000 inference.GRPCInferenceService/ModelMetadata


