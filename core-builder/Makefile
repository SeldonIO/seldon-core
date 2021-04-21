DOCKER_IMAGE_NAME=seldonio/core-builder
DOCKER_IMAGE_VERSION=0.20

build_docker_image:
	cp ../testing/scripts/dev_requirements.txt .
	docker build --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .
push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)
