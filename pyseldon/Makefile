DOCKER_IMAGE_NAME=seldonio/pyseldon2
DOCKER_IMAGE_VERSION=3.1

install:
	python setup.py install

test:
	nosetests -s -v seldon

build_docker_image:
	docker build --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .

push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)

update_protos:
	cp ../proto/prediction.proto ./seldon/proto
	cp ../proto/seldonengine.proto ./seldon/proto

build_protos:
	python -m grpc.tools.protoc -I./seldon/proto --python_out=./seldon/proto --grpc_python_out=./seldon/proto ./seldon/proto/prediction.proto
	python -m grpc.tools.protoc -I./seldon/proto --python_out=./seldon/proto --grpc_python_out=./seldon/proto ./seldon/proto/seldonengine.proto
