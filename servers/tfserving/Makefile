VERSION=latest
IMAGE_NAME_BASE=tfserving
IMG ?= seldonio/${IMAGE_NAME_BASE}:${VERSION}
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}


docker-build-redhat: 
	docker build . -f Dockerfile.redhat --build-arg TF_SERVING_VERSION=2.1.0 -t ${IMG_REDHAT}
