CUSTOM_IMAGE_TAG ?= latest
DOCKERHUB_USERNAME ?= seldonio
# Image URL to use all building/pushing image targets
IMG ?= ${DOCKERHUB_USERNAME}/seldonv2-controller:${CUSTOM_IMAGE_TAG}
IMG_CLI ?= ${DOCKERHUB_USERNAME}/seldon-cli:${CUSTOM_IMAGE_TAG}
AGENT_IMG ?= ${DOCKERHUB_USERNAME}/seldon-agent:${CUSTOM_IMAGE_TAG}
RCLONE_IMG ?= ${DOCKERHUB_USERNAME}/seldon-rclone:${CUSTOM_IMAGE_TAG}
MLSERVER_IMG ?= seldonio/mlserver:1.7.1
TRITON_IMG ?= nvcr.io/nvidia/tritonserver:23.03-py3
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23
CONTROLLER_GEN_VERSION = v0.18.0
KUSTOMIZE_VERSION = v5.2.1
SETUP_ENVTEST_VERSION = release-0.21
KIND_NAME=seldon
NAMESPACE ?= seldon-mesh

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
#SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=v2-manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	sed -i '/namespace: seldon-mesh/d' config/rbac/role.yaml # remove namespace added by controller-gen
	cp config/rbac/role.yaml config/rbac/namespace_role.yaml
	sed -i 's/ClusterRole/Role/' config/rbac/namespace_role.yaml
	cp config/rbac/role_binding.yaml config/rbac/namespace_role_binding.yaml
	sed -i 's/ClusterRole/Role/' config/rbac/namespace_role_binding.yaml
	sed -i 's/-{{[[:space:]]*\.Release\.Namespace[[:space:]]*}}//' config/rbac/namespace_role_binding.yaml

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.GOLANGCILINT_VERSION := v1.64.8
.GOLANGCILINT_PATH := $(shell go env GOPATH)/bin/golangci-lint-versions/$(.GOLANGCILINT_VERSION)

${.GOLANGCILINT_PATH}/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
			| sh -s -- -b ${.GOLANGCILINT_PATH} ${.GOLANGCILINT_VERSION}

.PHONY: lint
lint: ${.GOLANGCILINT_PATH}/golangci-lint
	${.GOLANGCILINT_PATH}/golangci-lint run --fix

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: clean-apis
	rm -rf apis-TEMP

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -trimpath -ldflags="-w" -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

build-seldon: generate fmt vet
	go build -trimpath -ldflags="-w" -o bin/seldon -v ./cmd/seldon

build-seldon-arm: generate fmt vet
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-w" -tags dynamic -o bin/seldon -v ./cmd/seldon

build-seldon-docgen:
	go build -trimpath -ldflags="-w" -o bin/seldon-gendocs -v ./cmd/seldon/docs

generate-cli-docs:
	./bin/seldon-gendocs --out ../docs/source/contents/cli/docs

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} -f Dockerfile ..

.PHONY: docker-build-cli
docker-build-cli: ## Build docker image with seldon cli.
	docker build -t ${IMG_CLI} -f Dockerfile.cli ..

.PHONY: docker-build-and-push-prod
docker-build-and-push-prod: ## Build operator docker image and push, including sbom.
	docker buildx build --provenance=true -t ${IMG} --attest type=sbom,generator=docker/scout-sbom-indexer:latest --push -f Dockerfile ..

.PHONY: docker-build-and-push-prod-cli
docker-build-and-push-prod-cli: ## Build seldon cli docker image and push, including sbom.
	docker buildx build --provenance=true -t ${IMG_CLI} --attest type=sbom,generator=docker/scout-sbom-indexer:latest --push -f Dockerfile.cli ..

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-push-cli
docker-push-cli: ## Push docker image with seldon cli.
	docker push ${IMG_CLI}

.PHONY: kind-image-load
kind-image-load:
	kind load -v 3 docker-image ${IMG} --name ${KIND_NAME}

.PHONY: kind-image-load-cli
kind-image-load-cli:
	kind load -v 3 docker-image ${IMG_CLI} --name ${KIND_NAME}

.PHONY: kind-image-load-all
kind-image-load-all: kind-image-load kind-image-load-cli

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = true
endif

.PHONY: deploy-seldon-cli
deploy-seldon-cli:
	cd config/cli && kustomize edit set image seldon-cli=${IMG_CLI} && kustomize edit set namespace ${NAMESPACE}
	kustomize build config/cli| kubectl apply -f -

.PHONY: undeploy-seldon-cli
undeploy-seldon-cli:
	kustomize build config/cli| kubectl delete -f -

# CRD is current too big for apply due to annotations size limit
.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl create -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

# CRD is current too big for apply due to annotations size limit
.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/crd | kubectl create -f -
	$(KUSTOMIZE) build config/default | kubectl create -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-server-configuration
deploy-server-configuration:
	cd config/serverconfigs && $(KUSTOMIZE) edit set image agent=${AGENT_IMG}
	cd config/serverconfigs && $(KUSTOMIZE) edit set image rclone=${RCLONE_IMG}
	cd config/serverconfigs && $(KUSTOMIZE) edit set image mlserver=${MLSERVER_IMG}
	cd config/serverconfigs && $(KUSTOMIZE) edit set image triton=${TRITON_IMG}
	$(KUSTOMIZE) build config/serverconfigs | kubectl create -f -

.PHONY: undeploy-server-configuration
undeploy-server-configuration:
	$(KUSTOMIZE) build config/serverconfigs | kubectl delete --ignore-not-found=$(ignore-not-found) -f -


.PHONY: deploy-servers
deploy-servers:
	$(KUSTOMIZE) build config/servers | kubectl create -f -

.PHONY: undeploy-servers
undeploy-servers:
	$(KUSTOMIZE) build config/servers | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-all
deploy-all: deploy deploy-server-configuration deploy-servers

.PHONY: undeploy-all
undeploy-all: undeploy-servers undeploy-server-configuration undeploy

build-push-deploy: build docker-build docker-push deploy-all

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION))

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION))

ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@$(SETUP_ENVTEST_VERSION))

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

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
