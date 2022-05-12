DOCKER_REPO ?= seldonio
IMAGE_NAME ?= hodometer
IMAGE_TAG ?= latest

HODOMETER_IMG = ${DOCKER_REPO}/${IMAGE_NAME}:${IMAGE_TAG}

GO_ENV = GO111MODULE=on CGO_ENABLED=0

# Build information
VERSION_PACKAGE = $(shell sed -n 's/^module\s\+//p' go.mod)/pkg
BUILD_VERSION = $(shell cat ./version.txt)
BUILD_TIME = $(shell date -u -Iseconds)
GIT_BRANCH = $(shell git branch --show-current)
GIT_COMMIT = $(shell git rev-parse --short=16 HEAD)
RELEASE_TYPE = pre-release

################################################################################

.PHONY: lint
lint:
	gofmt -w cmd
	gofmt -w pkg
	golangci-lint run --fix

.PHONY: build
build: test
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
				./cmd

.PHONY: test
test:
	$(GO_ENV) go test ./cmd ./pkg

.PHONY: clean
clean:
	rm bin/*

.PHONY: copy-apis
copy-apis:
	rm -rf apis-TEMP
	cp -r ../apis/mlops/scheduler apis-TEMP

.PHONY: build-docker
build-docker: copy-apis
	docker build -t ${HODOMETER_IMG} -f Dockerfile .
