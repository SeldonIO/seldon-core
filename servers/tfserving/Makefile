SHELL := /bin/bash
VERSION := $(shell cat ../../version.txt)
TF_SERVING_VERSION=2.1.0

IMAGE_NAME_BASE=tfserving
IMAGE_NAME ?= seldonio/${IMAGE_NAME_BASE}-ubi8


docker-build-redhat:
	docker build . -f Dockerfile.redhat --build-arg TF_SERVING_VERSION=${TF_SERVING_VERSION} -t ${IMAGE_NAME}:${VERSION}

# password can be found at: https://connect.redhat.com/projects/5ea822fdd0743cfaca95fc76/overview
project=5ea822fdd0743cfaca95fc76
redhat-image-scan: docker-build-redhat
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_tfserver} | docker login -u redhat-isv-containers+${project}-robot quay.io --password-stdin
	docker tag ${IMAGE_NAME}:${VERSION} quay.io/redhat-isv-containers/${project}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit
