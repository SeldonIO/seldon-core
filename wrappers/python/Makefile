build_docker_image:
	./build_image.sh

push_to_registry:
	./push_image.sh

build_proto:
	cp ../../proto/prediction.proto proto
	python -m grpc.tools.protoc -I./ --python_out=./ --grpc_python_out=./ ./proto/prediction.proto

seldon.json:
	cp ../../openapi/wrapper.oas3.json seldon.json

test:
	PYTHONPATH=. pytest 

clean:
	rm -f proto/prediction.proto
	rm -f proto/prediction_pb2_grpc.py
	rm -f proto/prediction_pb2.py
