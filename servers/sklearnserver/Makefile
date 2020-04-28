VERSION := $(shell cat ../../version.txt)
IMAGE_NAME_BASE=sklearnserver
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}
KIND_NAME ?= kind

build_rest:
	s2i build -E environment_rest ./sklearnserver seldonio/seldon-core-s2i-python37:${VERSION} ${IMAGE_BASE}_rest:${VERSION}

build_%:
	s2i build \
		-E environment_$* \
		./sklearnserver \
		seldonio/seldon-core-s2i-python37:${VERSION} \
		${IMAGE_BASE}_$*:${VERSION}

push_%:
	docker push ${IMAGE_BASE}_$*:${VERSION}

kind_load_%:
	kind load -v 3 docker-image ${IMAGE_BASE}_$*:${VERSION} --name ${KIND_NAME}

.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: build_all
build_all: build_rest build_grpc

.PHONY: kind_load_all
kind_load_all: kind_load_rest kind_load_grpc

#
# Redhat
#

IMG_VERSION_REDHAT_REST ?= ${IMAGE_NAME_BASE}_rest-ubi8:${VERSION}
IMG_REDHAT_REST ?= seldonio/${IMG_VERSION_REDHAT_REST}
IMG_VERSION_REDHAT_GRPC ?= ${IMAGE_NAME_BASE}_grpc-ubi8:${VERSION}
IMG_REDHAT_GRPC ?= seldonio/${IMG_VERSION_REDHAT_GRPC}

docker_build_redhat_rest:
	s2i build -E environment_rest ./sklearnserver seldonio/seldon-core-s2i-python37-ubi8:${VERSION} ${IMAGE_BASE}_rest-ubi8:${VERSION}

docker_build_redhat_grpc:
	s2i build -E environment_grpc ./sklearnserver seldonio/seldon-core-s2i-python37-ubi8:${VERSION} ${IMAGE_BASE}_grpc-ubi8:${VERSION}

# password can be found at: https://connect.redhat.com/project/4060551/view
redhat-image-scan_rest: 
	docker login -u unused scan.connect.redhat.com
	docker tag ${IMG_REDHAT_REST} scan.connect.redhat.com/ospid-5c2f945c-69fa-4de4-8574-e9a61f69d69a/${IMG_VERSION_REDHAT_REST}
	docker push scan.connect.redhat.com/ospid-5c2f945c-69fa-4de4-8574-e9a61f69d69a/${IMG_VERSION_REDHAT_REST}

# password can be found at: https://connect.redhat.com/project/4061631/view
redhat-image-scan_grpc: 
	docker login -u unused scan.connect.redhat.com
	docker tag ${IMG_REDHAT_GRPC} scan.connect.redhat.com/ospid-d75d8767-c689-4368-9859-4ec4f0062a3a/${IMG_VERSION_REDHAT_GRPC}
	docker push scan.connect.redhat.com/ospid-d75d8767-c689-4368-9859-4ec4f0062a3a/${IMG_VERSION_REDHAT_GRPC}

kind-image-install-redhat-rest: 
	kind load -v 3 docker-image ${IMG_REDHAT_REST}

kind-image-install-redhat-grpc: 
	kind load -v 3 docker-image ${IMG_REDHAT_GRPC}
