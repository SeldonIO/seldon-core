DOCKER_IMAGE_NAME=seldonio/python-builder
DOCKER_IMAGE_VERSION=0.4

build_docker_image:
	docker build --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .
	docker tag $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) $(DOCKER_IMAGE_NAME):latest
push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)
	docker push $(DOCKER_IMAGE_NAME):latest


test:
	docker run --rm -it $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) bash

check_docker_hub_tags:
	@echo "use: https://hub.docker.com/r/${DOCKER_IMAGE_NAME}/tags/"

