SHELL=/bin/bash


compile_proto:
	python -m grpc.tools.protoc -I ../../proto --python_out=src/ --grpc_python_out=src/ ../../proto/prediction.proto
	python -m grpc.tools.protoc -I ../../proto --python_out=src/ --grpc_python_out=src/ ../../proto/seldonengine.proto




