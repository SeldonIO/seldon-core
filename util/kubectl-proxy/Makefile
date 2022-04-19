DOCKER_IMAGE_NAME=seldonio/kubectl-proxy
KUBECTL_VERSION=v1.5.1
DOCKER_IMAGE_VERSION=$(KUBECTL_VERSION)

build_docker_image:
	docker build --build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .

push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)

