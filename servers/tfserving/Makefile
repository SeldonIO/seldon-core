SHELL := /bin/bash
VERSION=2.1.0
IMAGE_NAME_BASE=tfserving
IMG ?= seldonio/${IMAGE_NAME_BASE}:${VERSION}
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}


docker-build-redhat: 
	docker build . -f Dockerfile.redhat --build-arg TF_SERVING_VERSION=${VERSION} -t ${IMG_REDHAT}

# password can be found at: https://connect.redhat.com/project/4098071/view
redhat-image-scan: docker-build-redhat
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_tfserver} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-8714259c-f9bd-4ee9-9b6e-06f54354746e/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-8714259c-f9bd-4ee9-9b6e-06f54354746e/${IMG_VERSION_REDHAT}
