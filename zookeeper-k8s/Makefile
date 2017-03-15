DOCKER_IMAGE_NAME=seldonio/zookeeper-k8s
DOCKER_IMAGE_VERSION=1.0

build_docker_image:
	docker build --force-rm=true -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION) .

push_to_registry:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_VERSION)

