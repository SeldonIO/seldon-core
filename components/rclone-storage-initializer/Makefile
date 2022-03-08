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


docker-build-redhat:
	docker build --file=Dockerfile.redhat --force-rm=true -t seldonio/${IMAGE}-ubi8:${VERSION} .

docker-push-redhat:
	docker push seldonio/${IMAGE}-ubi8:${VERSION}

kind-load-redhat:
	kind load docker-image seldonio/${IMAGE}-ubi8:${VERSION} --name ${KIND_NAME}


scan=ospid-622b6fcf8a65f13d3bd4172f
redhat-image-scan:
	# docker pull seldonio/${IMAGE}-ubi8:${VERSION}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_rclone_storage_initializer} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag seldonio/${IMAGE}-ubi8:${VERSION} scan.connect.redhat.com/${scan}/${IMAGE}:${VERSION}
	docker push scan.connect.redhat.com/${scan}/${IMAGE}:${VERSION}
