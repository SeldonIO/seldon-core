build_docker_image:
	./build_image.sh

push_to_registry:
	./push_image.sh

build_proto:
	cp ../../proto/prediction.proto proto
	cp -vr ../../proto/tensorflow/tensorflow proto
	cd proto && python -m grpc.tools.protoc -I./ --python_out=./ --grpc_python_out=./ ./prediction.proto

build_fbs:
	cp ../../fbs/prediction.fbs fbs
	flatc --python -o fbs fbs/prediction.fbs

seldon.json:
	cp ../../openapi/wrapper.oas3.json seldon.json

test:
	PYTHONPATH=. pytest -W ignore

clean:
	rm -f proto/prediction.proto
	rm -f proto/prediction_pb2_grpc.py
	rm -f proto/prediction_pb2.py
	rm -f fbs/*.py*
	rm -f fbs/prediction.fbs
