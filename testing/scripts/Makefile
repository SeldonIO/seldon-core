
build_protos:
	cp ../../proto/prediction.proto ./proto
	cd ../../proto/tensorflow && make create_protos
	cp -vr ../../proto/tensorflow/tensorflow .
	python -m grpc.tools.protoc -I. --python_out=. --grpc_python_out=. ./proto/prediction.proto
	cd ../../notebooks && make build_protos

#requires mvn, helm, ks, python, s2i and an accessible k8s cluster
#tests will create namespaces, deploy seldon core and create examples

.PHONY: test
test: build_protos
	pytest -s -W ignore > run.log 2>&1

.PHONY: clean
clean:
	rm -f proto/prediction*
