update_protos:
	cp ../proto/prediction.proto ./proto

build_protos:
	python -m grpc.tools.protoc -I./proto --python_out=./proto --grpc_python_out=./proto ./proto/prediction.proto

