VERSION=0.2
IMAGE_BASE=seldonio/sklearnserver

build_rest:
	s2i build -E environment_rest ./sklearnserver seldonio/seldon-core-s2i-python37:0.12-SNAPSHOT ${IMAGE_BASE}_rest:${VERSION}

push_rest:
	docker push ${IMAGE_BASE}_rest:${VERSION}

build_grpc:
	s2i build -E environment_grpc ./sklearnserver seldonio/seldon-core-s2i-python37:0.12-SNAPSHOT ${IMAGE_BASE}_grpc:${VERSION}

push_grpc:
	docker push ${IMAGE_BASE}_grpc:${VERSION}


.PHONY: push_all
push_all: push_rest push_grpc

.PHONY: build_all
build_all: build_rest build_grpc
