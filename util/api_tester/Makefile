
.PHONY: update_protos
update_protos:
	mkdir -p ./proto
	touch ./proto/__init__.py
	cp ../../proto/prediction.proto ./proto
	cd ../../proto/tensorflow && make create_protos
	cp -vr ../../proto/tensorflow/tensorflow .

.PHONY: build_protos
build_protos: clean update_protos
	python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. ./proto/prediction.proto

.PHONY: clean
clean:
	rm -f proto/prediction*.py
	rm -f proto/prediction.proto
	rm -rf proto/__pycache__

