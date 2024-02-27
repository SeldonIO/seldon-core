DOCKER_REPO ?= seldonio
IMAGE_NAME ?= seldon-hodometer
IMAGE_TAG ?= latest

HODOMETER_IMG = ${DOCKER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}
RECEIVER_IMG = ${DOCKER_REPO}/${IMAGE_NAME}-receiver:${IMAGE_TAG}

GO_ENV = GO111MODULE=on CGO_ENABLED=0

VERBOSE ?=

KIND_NAME=seldon

# Build information
VERSION_PACKAGE = $(shell sed -n 's/^module\s\+//p' go.mod)/pkg/hodometer
BUILD_VERSION ?= $(shell cat ./version.txt)
BUILD_TIME = $(shell date -u -Iseconds)
GIT_BRANCH ?= $(shell git branch --show-current)
GIT_COMMIT ?= $(shell git rev-parse --short=16 HEAD)
RELEASE_TYPE ?= pre-release

################################################################################

.GOLANGCILINT_VERSION := v1.51.2
.GOLANGCILINT_PATH := $(shell go env GOPATH)/bin/golangci-lint/$(.GOLANGCILINT_VERSION)

${.GOLANGCILINT_PATH}/golangci-lint: 
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
			| sh -s -- -b ${.GOLANGCILINT_PATH} ${.GOLANGCILINT_VERSION}

.PHONY: lint
lint: ${.GOLANGCILINT_PATH}/golangci-lint
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
build-hodometer-local: copy-components build-hodometer

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

.PHONY: build-hodometer-docker
build-hodometer-docker: 
	docker build \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg GIT_BRANCH=$(GIT_BRANCH) \
		--build-arg RELEASE_TYPE=$(RELEASE_TYPE) \
		--build-arg BUILD_VERSION=$(BUILD_VERSION) \
		-t ${HODOMETER_IMG} \
		-f Dockerfile.hodometer ..

.PHONY: build-receiver-docker
build-receiver-docker:
	docker build -t ${RECEIVER_IMG} -f Dockerfile.receiver ..

.PHONY: push-hodometer-docker
push-hodometer-docker:
	docker push ${HODOMETER_IMG}

.PHONY: kind-image-install-hodometer
kind-image-install-hodometer:
	kind load -v 3 docker-image ${HODOMETER_IMG} --name ${KIND_NAME}

.PHONY: licenses/dep.txt
licenses/dep.txt:
	go list -m all | cut -d ' ' -f 1 > licenses/dep.txt

.PHONY: licenses
licenses: licenses/dep.txt
	# NOTE: You need to create a file in ~/.github_api_token with a GitHub token.
	get-github-repo \
		-o licenses/repo.txt \
		--manual-dep-repo-mapping ../licenses/dep_repo.manual.csv \
		licenses/dep.txt
	get-github-license-info -o licenses/license_info.csv licenses/repo.txt --branch-refs ../licenses/branch_refs.manual.csv
	python -m 'patch_additional_license_info' \
		licenses/license_info.csv \
		../licenses/additional_license_info.csv
	concatenate-license -o licenses/license.txt licenses/license_info.csv
