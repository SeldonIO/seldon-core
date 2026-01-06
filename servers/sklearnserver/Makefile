SHELL := /bin/bash
VERSION ?= $(shell cat ../../version.txt)

DOCKER_REGISTRY ?= seldonio

IMAGE_NAME_BASE = sklearnserver
IMAGE_NAME = ${DOCKER_REGISTRY}/${IMAGE_NAME_BASE}
IMAGE_NAME_PROD_BASE=${IMAGE_NAME}-prod-base
KIND_NAME ?= kind

BASE_IMAGE = ${DOCKER_REGISTRY}/seldon-core-s2i-python312:${VERSION}

docker-build:
	s2i build \
		-E environment \
		./sklearnserver \
		${BASE_IMAGE} \
		${IMAGE_NAME}:${VERSION}

# Build and push s2i base image to the private repository in dockerhub
# After it's used as a base image for building the final image with SBOM attached
docker-push-prod-base: docker-build
	docker tag ${IMAGE_NAME}:${VERSION} ${IMAGE_NAME_PROD_BASE}:${VERSION}
	docker push ${IMAGE_NAME_PROD_BASE}:${VERSION}

docker-build-and-push-prod: docker-push-prod-base
	docker buildx build --provenance=true --attest type=sbom,generator=docker/scout-sbom-indexer:latest -f Dockerfile.s2i --build-arg BASE_IMAGE=${IMAGE_NAME_PROD_BASE} --build-arg BASE_TAG=${VERSION} -t ${IMAGE_NAME}:${VERSION} . --push

docker-push:
	docker push ${IMAGE_NAME}:${VERSION}

kind-load: docker-build
	kind load -v 3 docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_NAME}

# password can be found at: https://connect.redhat.com/projects/5fb94ae7888dca360c9231e1/overview
project=5fb94ae7888dca360c9231e1
redhat-image-scan:
	docker pull ${IMAGE_NAME}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_sklearnserver} | docker login -u redhat-isv-containers+${project}-robot quay.io --password-stdin
	docker tag ${IMAGE_NAME}:${VERSION} quay.io/redhat-isv-containers/${project}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit
