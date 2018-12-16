.PHONY: update_protos
update_protos:
	cp ../../proto/prediction.proto ./proto
	cd ../../proto/tensorflow && make create_protos
	cp -vr ../../proto/tensorflow/tensorflow .

.PHONY: build_protos
build_protos: clean update_protos
	python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. ./proto/prediction.proto

.PHONY: build_fbs
build_fbs: 
	flatc --python -o fbs ../../fbs/prediction.fbs
	touch fbs/__init__.py

.PHONY: clean
clean:
	rm -f proto/prediction*.py
	rm -f proto/prediction.proto
	rm -rf proto/__pycache__
	rm -f fbs/*.py
	rm -rf fbs/__pycache__


