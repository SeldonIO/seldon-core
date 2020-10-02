DOCKER_REPOSITORY ?= seldonio
IMAGE_NAME ?= distributed-roberta
IMAGE_VERSION ?= 0.1

KIND_NAME ?= kind

# Image build and push
build-image:
	docker build . -t ${DOCKER_REPOSITORY}/${IMAGE_NAME}:${IMAGE_VERSION}

push-image:
	docker push ${DOCKER_REPOSITORY}/${IMAGE_NAME}:${IMAGE_VERSION}

kind-load:
	kind load docker-image ${DOCKER_REPOSITORY}/${IMAGE_NAME}:${IMAGE_VERSION}
