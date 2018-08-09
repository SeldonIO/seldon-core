.PHONY: update_protos
update_protos:
	cp ../../proto/prediction.proto ./proto

.PHONY: build_protos
build_protos: clean update_protos
	python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. ./proto/prediction.proto

.PHONY: build_fbs
build_fbs: 
	flatc --python -o fbs ../../fbs/prediction.fbs
	mv fbs/io/seldon/seldonmarketpredictor/fbs/prediction/* fbs
	rm -rf fbs/io

.PHONY: clean
clean:
	rm -f proto/prediction*.py
	rm -f proto/prediction.proto
	rm -rf proto/__pycache__
	rm -f fbs/*.py
	rm -rf fbs/__pycache__


