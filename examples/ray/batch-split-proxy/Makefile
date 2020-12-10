DOCKER_REPOSITORY ?= seldonio
IMAGE_NAME ?= distributed-roberta
IMAGE_VERSION ?= 0.1

KIND_NAME ?= kind

download:
	mkdir -p training/outputs
	curl -X GET -o training/outputs/pytorch_model.bin "https://storage.googleapis.com/seldon-models/pytorch/moviesentiment_roberta/pytorch_model_v2.bin"


# Image build and push
build-image:
	docker build . -t ${DOCKER_REPOSITORY}/${IMAGE_NAME}:${IMAGE_VERSION}

push-image:
	docker push ${DOCKER_REPOSITORY}/${IMAGE_NAME}:${IMAGE_VERSION}

kind-load:
	kind load docker-image ${DOCKER_REPOSITORY}/${IMAGE_NAME}:${IMAGE_VERSION}
