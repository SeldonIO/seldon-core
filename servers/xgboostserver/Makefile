VERSION=0.3-SNAPSHOT
IMAGE_BASE=seldonio/xgboostserver
S2I_IMAGE_VERSION=0.19-SNAPSHOT

build_%:
	s2i build \
		-E environment_$* \
		./xgboostserver \
		seldonio/seldon-core-s2i-python37:${S2I_IMAGE_VERSION} \
		${IMAGE_BASE}_$*:${VERSION}

push_%:
	docker push ${IMAGE_BASE}_$*:${VERSION}

kind_load_%:
	kind load -v 3 docker-image ${IMAGE_BASE}_$*:${VERSION}

.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: build_all
build_all: build_rest build_grpc

.PHONY: kind_load_all
kind_load_all: kind_load_rest kind_load_grpc
