DOCKER_REPO ?= seldonio
IMAGE_NAME ?= seldon-hodometer
IMAGE_TAG ?= latest

HODOMETER_IMG = ${DOCKER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}
RECEIVER_IMG = ${DOCKER_REPO}/${IMAGE_NAME}-receiver:${IMAGE_TAG}

GO_ENV = GO111MODULE=on CGO_ENABLED=0

VERBOSE ?=

KIND_NAME=ansible

# Build information
VERSION_PACKAGE = $(shell sed -n 's/^module\s\+//p' go.mod)/pkg/hodometer
BUILD_VERSION ?= $(shell cat ./version.txt)
BUILD_TIME = $(shell date -u -Iseconds)
GIT_BRANCH ?= $(shell git branch --show-current)
GIT_COMMIT ?= $(shell git rev-parse --short=16 HEAD)
RELEASE_TYPE ?= pre-release

################################################################################

.GOLANGCILINT_VERSION := v1.48.0
.GOLANGCILINT_PATH := $(shell go env GOPATH)/bin/golangci-lint/$(.GOLANGCILINT_VERSION)

.PHONY: .install-golangci-lint
.ONESHELL: .install-golangci-lint
.install-golangci-lint:
	@installed=false
	if [ -e "${.GOLANGCILINT_PATH}/golangci-lint" ]; then
		installed=true
	fi
	if ! $${installed}; then
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
			| sh -s -- -b ${.GOLANGCILINT_PATH} ${.GOLANGCILINT_VERSION}
	fi

.PHONY: lint
lint: .install-golangci-lint
	gofmt -w cmd
	gofmt -w pkg
	${.GOLANGCILINT_PATH}/golangci-lint run --fix

.PHONY: build-hodometer
build-hodometer: test-hodometer 
	$(GO_ENV) \
				go build \
				-o bin/hodometer \
				-ldflags=" \
				-X '$(VERSION_PACKAGE).BuildVersion=$(BUILD_VERSION)' \
				-X '$(VERSION_PACKAGE).BuildTime=$(BUILD_TIME)' \
				-X '$(VERSION_PACKAGE).GitBranch=$(GIT_BRANCH)' \
				-X '$(VERSION_PACKAGE).GitCommit=$(GIT_COMMIT)' \
				-X '$(VERSION_PACKAGE).ReleaseType=$(RELEASE_TYPE)' \
				" \
				./cmd/hodometer/
.PHONY: build-hodometer-local
build-hodometer-local: copy-apis copy-components build-hodometer

.PHONY: build-receiver
build-receiver: test-receiver
	$(GO_ENV) go build -o bin/receiver ./cmd/receiver

.PHONY: build
build: build-hodometer build-receiver

.PHONY: test-hodometer
test-hodometer:
	$(GO_ENV) go test $(VERBOSE) ./cmd/hodometer/... ./pkg/hodometer/...

.PHONY: test-receiver
test-receiver:
	$(GO_ENV) go test $(VERBOSE) ./cmd/receiver/... ./pkg/receiver/...

.PHONY: test
test: test-hodometer test-receiver

.PHONY: clean
clean:
	rm bin/*

.PHONY: copy-apis
copy-apis:
	rm -rf apis-TEMP
	cp -r ../apis/mlops/scheduler apis-TEMP

.PHONY: copy-components
copy-components:
	rm -rf components-TEMP
	cp -r ../components components-TEMP

.PHONY: build-hodometer-docker
build-hodometer-docker: copy-apis copy-components
	docker build \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg GIT_BRANCH=$(GIT_BRANCH) \
		--build-arg RELEASE_TYPE=$(RELEASE_TYPE) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		-t ${HODOMETER_IMG} \
		-f Dockerfile.hodometer .

.PHONY: build-receiver-docker
build-receiver-docker:
	docker build -t ${RECEIVER_IMG} -f Dockerfile.receiver .

.PHONY: push-hodometer-docker
push-hodometer-docker:
	docker push ${HODOMETER_IMG}

.PHONY: kind-image-install-hodometer
kind-image-install-hodometer:
	kind load -v 3 docker-image ${HODOMETER_IMG} --name ${KIND_NAME}
