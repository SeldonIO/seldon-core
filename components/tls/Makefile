GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

.PHONY: test
test:
	go test ./pkg/... -coverprofile cover.out

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
	gofmt -w pkg
	${.GOLANGCILINT_PATH}/golangci-lint run --fix

.PHONY: build
build: test
	go build ./pkg/...
