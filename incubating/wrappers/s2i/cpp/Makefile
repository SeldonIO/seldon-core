BASE_IMAGE_TAG := 1.5.1
BASE_IMAGE_NAME := docker.io/seldonio/seldon-core-s2i-python37-ubi8
SHELL := /bin/bash
SELDON_CORE_DIR = ../../../../
VERSION := 0.0.1
IMAGE_REGISTRY = docker.io/seldonio
IMAGE_NAME = ${IMAGE_REGISTRY}/s2i-cpp-build:${VERSION}

cmake:
	cmake ./ -Bbuild \
		-DSELDON_OPT_BUILD_TESTS=1 \
		-DSELDON_OPT_BUILD_STATIC=1 \
		-DSELDON_OPT_BUILD_PROTO=1 \
		-DSELDON_OPT_INSTALL=1 \
		-DSELDON_OPT_CLONE_PYBIND11=1 

cmake-build:
	make -C build/ -j$(shell nproc)

cmake-test: cmake-build
	./build/test/seldon-test

cmake-clean:
	rm -rf build/

proto-build:
	mkdir -p core/
	protoc -I=${SELDON_CORE_DIR}/proto -I=${SELDON_CORE_DIR}/proto/tensorflow \
		--cpp_out=core/ \
		${SELDON_CORE_DIR}/proto/prediction.proto
	mv core/prediction.pb.h src/include/prediction.pb.h
	mv core/prediction.pb.cc src/proto/prediction.pb.cc
	protoc -I=${SELDON_CORE_DIR}/proto/tensorflow \
		--cpp_out=core/ \
		${SELDON_CORE_DIR}/proto/tensorflow/tensorflow/core/framework/tensor.proto
	protoc -I=${SELDON_CORE_DIR}/proto/tensorflow \
		--cpp_out=core/ \
		${SELDON_CORE_DIR}/proto/tensorflow/tensorflow/core/framework/types.proto
	protoc -I=${SELDON_CORE_DIR}/proto/tensorflow \
		--cpp_out=core/ \
		${SELDON_CORE_DIR}/proto/tensorflow/tensorflow/core/framework/tensor_shape.proto
	protoc -I=${SELDON_CORE_DIR}/proto/tensorflow \
		--cpp_out=core/ \
		${SELDON_CORE_DIR}/proto/tensorflow/tensorflow/core/framework/resource_handle.proto
	mkdir -p core/src/include/tensorflow/core/framework
	mv core/tensorflow/core/framework/*.h src/include/tensorflow/core/framework
	mv core/tensorflow/core/framework/*.cc src/proto/
	rm -rf core/

clone-pybind:
	git clone --depth 1 --branch v2.6.1 https://github.com/pybind/pybind11 || echo "Pybind not cloned"

docker-build:
	docker build -f Dockerfile \
		-t $(IMAGE_NAME) \
		--build-arg BASE_IMAGE_NAME=${BASE_IMAGE_NAME} \
		--build-arg BASE_IMAGE_TAG=${BASE_IMAGE_TAG} \
		./

docker-push:
	docker push $(IMAGE_NAME)

