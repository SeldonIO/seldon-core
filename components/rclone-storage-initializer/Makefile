SHELL := /bin/bash
IMAGE := rclone-storage-initializer
VERSION := $(shell cat ../../version.txt)

KIND_NAME ?= kind

docker-build:
	docker build --file=Dockerfile --force-rm=true -t seldonio/${IMAGE}:${VERSION} .

docker-push:
	docker push seldonio/${IMAGE}:${VERSION}

kind-load: docker-build
	kind load docker-image seldonio/${IMAGE}:${VERSION} --name ${KIND_NAME}

# password can be found at: https://connect.redhat.com/projects/622b6fcf8a65f13d3bd4172f/overview
project=622b6fcf8a65f13d3bd4172f
redhat-image-scan:
	docker pull seldonio/${IMAGE}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_rclone_storage_initializer} | docker login -u redhat-isv-containers+${project}-robot quay.io --password-stdin
	docker tag seldonio/${IMAGE}:${VERSION} quay.io/redhat-isv-containers/${project}:${VERSION}
	docker push quay.io/redhat-isv-containers/${project}:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		preflight check container quay.io/redhat-isv-containers/${project}:${VERSION} --docker-config=${HOME}/.docker/config.json --certification-project-id=${project} --pyxis-api-token=$${pyxis_api_token} --submit
