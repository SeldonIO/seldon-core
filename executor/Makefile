SHELL := /bin/bash
VERSION := $(shell cat ../version.txt)
# Image URL to use all building/pushing image targets
IMAGE_NAME_BASE=seldon-core-executor
IMG ?= seldonio/${IMAGE_NAME_BASE}:${VERSION}
IMG_VERSION_REDHAT ?= ${IMAGE_NAME_BASE}-ubi8:${VERSION}
IMG_REDHAT ?= seldonio/${IMG_VERSION_REDHAT}

EXECUTOR_FOLDERS ?= ./api/... ./predictor/... ./k8s/... ./logger/...

KIND_NAME ?= kind

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Run go fmt against code
fmt:
	go fmt ${EXECUTOR_FOLDERS}

# Run go vet against code
vet:
	go vet ${EXECUTOR_FOLDERS}


# Build manager binary
executor: copy_operator fmt vet
	go build -o executor cmd/executor/main.go


kafka-proxy: copy_operator fmt vet
	go build -o kafka-proxy cmd/proxy/main.go


.PHONY: copy_operator
copy_operator:
	rm -rf _operator
	cp -r ../operator _operator


.PHONY: copy_protos
copy_protos:
	cp -r ../proto/tensorflow/tensorflow/** proto/tensorflow


.PHONY: compile_seldon_proto
compile_seldon_proto:
	cp ../proto/prediction.proto api/grpc/seldon/proto
	cd api/grpc/seldon/proto && protoc -I. -I../../../../proto/ --go_out=paths=source_relative,plugins=grpc:. prediction.proto
	sed -i "s/package api/package proto/" api/grpc/seldon/proto/prediction.pb.go
	rm api/grpc/seldon/proto/prediction.proto

# https://github.com/tensorflow/serving/issues/1365#issuecomment-525351995
.PHONY: compile_tensorflow_proto
compile_tensorflow_proto:
	git clone -b r1.15 https://github.com/tensorflow/tensorflow.git
	git clone -b r1.14 https://github.com/tensorflow/serving.git
	go run protoc.go
	go mod edit -replace=github.com/tensorflow/tensorflow/tensorflow/go/core=./proto/tensorflow/core
	cd proto/tensorflow/core && go mod init github.com/tensorflow/tensorflow/tensorflow/go/core && cd -
	go build ./proto/tensorflow/serving


compile_kfserving_proto:
	#git clone https://github.com/NVIDIA/triton-inference-server
	mkdir -p api/grpc/kfserving/inference
	cp triton-inference-server/src/core/*.proto api/grpc/kfserving/inference
	cd api/grpc/kfserving/inference && protoc -I. --go_out=paths=source_relative,plugins=grpc:. grpc_service.proto && protoc -I. --go_out=paths=source_relative,plugins=grpc:. model_config.proto && protoc -I. --go_out=paths=source_relative,plugins=grpc:. server_status.proto
	#sed -i "s/package nvidia_inferenceserver/package proto/" api/grpc/kfserving/proto/grpc_service.pb.go
	#sed -i "s/package nvidia_inferenceserver/package proto/" api/grpc/kfserving/proto/model_config.pb.go
	#sed -i "s/package nvidia_inferenceserver/package proto/" api/grpc/kfserving/proto/server_status.pb.go


.PHONY: add_protos
add_protos:
	cd tensorflow && find ./tensorflow -name '*.proto' | cpio -pdm ../proto
	cd serving && find ./tensorflow_serving -name '*.proto' | cpio -pdm ../proto

# Run tests
test: copy_operator fmt vet
	go test ${EXECUTOR_FOLDERS} -coverprofile cover.out

copy_openapi_resources:
	mkdir -p api/rest/openapi/
	cp ../openapi/swagger-ui/* api/rest/openapi/
	cp ../openapi/engine.oas3.json api/rest/openapi/seldon.json

# Build the docker image
docker-build: test copy_openapi_resources
	docker build -f Dockerfile.executor -t ${IMG} .

# Build the docker image for Redhat
docker-build-redhat: test copy_openapi_resources
	docker build -f Dockerfile.executor.redhat -t ${IMG_REDHAT} .

# Push the docker image
docker-push:
	docker push ${IMG}

# Push the docker image
docker-push-redhat:
	docker push ${IMG_REDHAT}

# password can be found at: https://connect.redhat.com/project/3977851/view
redhat-image-scan:
	docker pull ${IMG_REDHAT}
	source ~/.config/seldon/seldon-core/redhat-image-passwords.sh && \
		echo $${rh_password_executor} | docker login -u unused scan.connect.redhat.com --password-stdin
	docker tag ${IMG_REDHAT} scan.connect.redhat.com/ospid-fffb6f6b-90b5-4f1d-be69-71baa8fb16cb/${IMG_VERSION_REDHAT}
	docker push scan.connect.redhat.com/ospid-fffb6f6b-90b5-4f1d-be69-71baa8fb16cb/${IMG_VERSION_REDHAT}

kind-image-install: docker-build
	kind load -v 3 docker-image ${IMG} --name ${KIND_NAME}

kind-image-install-redhat: docker-build-redhat
	kind load -v 3 docker-image ${IMG_REDHAT} --name ${KIND_NAME}


.PHONY: clean
clean:
	rm -rf vendor
	rm -rf tensorflow
	rm -rf serving
	rm -rf triton-inference-server

install-dev:
	# Tool to generate license info
	pip install \
		'git+https://github.com/kubeflow/testing#egg=go-license-tools&subdirectory=py/kubeflow/testing/go-license-tools'


.PHONY: licenses/dep.txt
licenses/dep.txt:
	go list -m all | cut -d ' ' -f 1 > licenses/dep.txt


.PHONY: licenses
licenses: licenses/dep.txt
	# NOTE: You need to create a file in ~/.github_api_token with a GitHub token.
	get-github-repo \
		-o licenses/repo.txt \
		--manual-dep-repo-mapping licenses/dep_repo.manual.csv \
		licenses/dep.txt
	get-github-license-info -o licenses/license_info.csv licenses/repo.txt
	python -m 'patch_additional_license_info' \
		licenses/license_info.csv \
		licenses/additional_license_info.csv
	concatenate-license -o licenses/license.txt licenses/license_info.csv


.PHONY: lint
lint: copy_operator licenses/dep.txt
	# Check if licenses have changed
	git \
		--no-pager diff \
		--exit-code \
		./licenses
