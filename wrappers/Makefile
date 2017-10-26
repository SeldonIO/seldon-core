update_protos:
	cp ../proto/prediction.proto ./python/proto
	cp ../proto/seldonengine.proto ./python/proto

build_protos:
	python -m grpc.tools.protoc -I./python/proto --python_out=./python/proto --grpc_python_out=./python/proto ./python/proto/prediction.proto
	python -m grpc.tools.protoc -I./python/proto --python_out=./python/proto --grpc_python_out=./python/proto ./python/proto/seldonengine.proto
