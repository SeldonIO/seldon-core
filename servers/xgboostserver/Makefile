VERSION=0.3-SNAPSHOT
IMAGE_BASE=seldonio/xgboostserver
S2I_IMAGE_VERSION=0.19-SNAPSHOT


build_rest:
	s2i build \
		-E environment_rest \
		./xgboostserver \
		seldonio/seldon-core-s2i-python37:${S2I_IMAGE_VERSION} \
		${IMAGE_BASE}_rest:${VERSION}

push_rest:
	docker push ${IMAGE_BASE}_rest:${VERSION}

build_grpc:
	s2i build \
		-E environment_grpc \
		./xgboostserver \
		seldonio/seldon-core-s2i-python37:${S2I_IMAGE_VERSION} \
		${IMAGE_BASE}_grpc:${VERSION}

push_grpc:
	docker push ${IMAGE_BASE}_grpc:${VERSION}


.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: build_all
build_all: build_rest build_grpc
