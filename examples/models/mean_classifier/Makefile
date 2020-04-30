VERSION := $(shell cat ../../../version.txt)
IMAGE_NAME_BASE=mock_classifier
IMAGE_BASE=seldonio/${IMAGE_NAME_BASE}


build_rest:
	s2i build -E environment_rest . seldonio/seldon-core-s2i-python36:${VERSION} ${IMAGE_BASE}_rest:${VERSION}

push_rest:
	docker push ${IMAGE_BASE}_rest:${VERSION}

build_grpc:
	s2i build -E environment_grpc . seldonio/seldon-core-s2i-python36:${VERSION} ${IMAGE_BASE}_grpc:${VERSION}

push_grpc:
	docker push ${IMAGE_BASE}_grpc:${VERSION}

run_rest_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && TRACING=1 JAEGER_AGENT_HOST=localhost JAEGER_AGENT_PORT=6831 JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 seldon-core-microservice --service-type MODEL MeanClassifier REST

run_grpc_local:
	export PREDICTIVE_UNIT_SERVICE_PORT=9000 && TRACING=1 JAEGER_AGENT_HOST=localhost JAEGER_AGENT_PORT=6831 JAEGER_SAMPLER_TYPE=const JAEGER_SAMPLER_PARAM=1 seldon-core-microservice --service-type MODEL MeanClassifier GRPC


kind_load_rest:
	kind load -v 3 docker-image ${IMAGE_BASE}_rest:${VERSION}

kind_load_grpc:
	kind load -v 3 docker-image ${IMAGE_BASE}_grpc:${VERSION}

#
# Redhat
#

IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}

docker-build-redhat:
	docker build . -f Dockerfile.redhat -t ${IMG_REDHAT}

# password can be found at: https://connect.redhat.com/project/4035711/view
redhat-image-scan:
	docker login -u unused scan.connect.redhat.com
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-c90fdfe6-d054-4598-baa8-7f7aac3ed63a/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-c90fdfe6-d054-4598-baa8-7f7aac3ed63a/${IMG_VERSION_REDHAT}
