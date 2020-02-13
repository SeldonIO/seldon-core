VERSION=0.3-SNAPSHOT
IMAGE_BASE=seldonio/mlflowserver

build_rest:
	s2i build -E environment_rest ./mlflowserver seldonio/seldon-core-s2i-python37:0.17-SNAPSHOT ${IMAGE_BASE}_rest:${VERSION}

push_rest:
	docker push ${IMAGE_BASE}_rest:${VERSION}

build_grpc:
	s2i build -E environment_grpc ./mlflowserver seldonio/seldon-core-s2i-python37:0.17-SNAPSHOT ${IMAGE_BASE}_grpc:${VERSION}

push_grpc:
	docker push ${IMAGE_BASE}_grpc:${VERSION}

.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: build_all
build_all: build_rest build_grpc
