update_protos:
	cp ../proto/prediction.proto ./python/proto

build_protos:
	python -m grpc.tools.protoc -I./python --python_out=./python --grpc_python_out=./python ./python/proto/prediction.proto

