SHELL:=/bin/bash

DOCKER_IMAGE_NAME=seldonio/locust-core
DOCKER_IMAGE_VERSION=0.8

build_proto:
	cp ../../proto/prediction.proto scripts/proto
	cd ../../proto/tensorflow && make create_protos
	cp -vr ../../proto/tensorflow/tensorflow scripts
	cd scripts && python -m grpc.tools.protoc -I./ --python_out=./ --grpc_python_out=./ ./proto/prediction.proto


build_docker_image: build_proto
	docker build --force-rm=true -t $(PRIVATE_REPO)$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .

push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)


