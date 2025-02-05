GO_LDFLAGS := -s -w $(patsubst %,-X %, $(GO_BUILD_VARS))

.PHONY: test
test:
	go test -v ./pkg/... -coverprofile cover.out

.PHONY: test-integration
test-integration:
	go test -v ./pkg/... -tags=integration

.GOLANGCILINT_VERSION := v1.57.2
.GOLANGCILINT_PATH := $(shell go env GOPATH)/bin/golangci-lint/$(.GOLANGCILINT_VERSION)

${.GOLANGCILINT_PATH}/golangci-lint: 
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
			| sh -s -- -b ${.GOLANGCILINT_PATH} ${.GOLANGCILINT_VERSION}

.PHONY: lint
lint: ${.GOLANGCILINT_PATH}/golangci-lint
	gofmt -w pkg
	${.GOLANGCILINT_PATH}/golangci-lint run --fix

.PHONY: licenses/dep.txt
licenses/dep.txt:
	go list -m all | cut -d ' ' -f 1 > licenses/dep.txt

.PHONY: licenses
licenses: licenses/dep.txt
	# NOTE: You need to create a file in ~/.github_api_token with a GitHub token.
	get-github-repo \
		-o licenses/repo.txt \
		--manual-dep-repo-mapping ../../licenses/dep_repo.manual.csv \
		licenses/dep.txt
	get-github-license-info -o licenses/license_info.csv licenses/repo.txt --branch-refs ../../licenses/branch_refs.manual.csv
	python -m 'patch_additional_license_info' \
		licenses/license_info.csv \
		../../licenses/additional_license_info.csv
	concatenate-license -o licenses/license.txt licenses/license_info.csv
