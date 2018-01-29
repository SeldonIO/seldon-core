SHELL:=/bin/bash

DOCKER_IMAGE_NAME=seldonio/locust-core
DOCKER_IMAGE_VERSION=0.4

build_docker_image: 
	docker build --force-rm=true -t $(PRIVATE_REPO)$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .

push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)


