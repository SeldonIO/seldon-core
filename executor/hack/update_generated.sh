#!/usr/bin/env bash
set -o nounset
set -o errexit
set -o pipefail
set -o xtrace

PKG_BASE=github.com/seldonio/seldon-core/executor
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel)}"
REPO_ROOT=${REPO_ROOT}/executor
cd "${REPO_ROOT}"

# enable modules and the proxy cache
export GO111MODULE="on"

# build the generators
BINDIR="${REPO_ROOT}/bin"
# use the tools module
cd "hack/tools"
go build -o "${BINDIR}/defaulter-gen" k8s.io/code-generator/cmd/defaulter-gen
go build -o "${BINDIR}/deepcopy-gen" k8s.io/code-generator/cmd/deepcopy-gen
go build -o "${BINDIR}/conversion-gen" k8s.io/code-generator/cmd/conversion-gen
go build -o "${BINDIR}/client-gen" k8s.io/code-generator/cmd/client-gen
go build -o "${BINDIR}/lister-gen" k8s.io/code-generator/cmd/lister-gen
go build -o "${BINDIR}/informer-gen" k8s.io/code-generator/cmd/informer-gen
# go back to the root
cd "${REPO_ROOT}"


# turn off module mode before running the generators
# https://github.com/kubernetes/code-generator/issues/69
# we also need to populate vendor
go mod vendor
export GO111MODULE="off"

# fake being in a gopath
FAKE_GOPATH="$(mktemp -d)"
#trap 'rm -rf ${FAKE_GOPATH}' EXIT

FAKE_REPOPATH="${FAKE_GOPATH}/src/${PKG_BASE}"
mkdir -p "$(dirname "${FAKE_REPOPATH}")" && ln -s "${REPO_ROOT}" "${FAKE_REPOPATH}"

export GOPATH="${FAKE_GOPATH}"
cd "${FAKE_REPOPATH}"

# run the generators
#"${BINDIR}/deepcopy-gen" -v 9 -i ./api/v1alpha2/ -O zz_generated_new.deepcopy --go-header-file hack/boilerplate.go.txt

OUTPUT_PKG=${FAKE_REPOPATH}/client 

"${BINDIR}/client-gen" -v 9 --input-base ${PKG_BASE}/api --clientset-name versioned -i ./api/machinelearning/v1alpha2/ --input machinelearning/v1alpha2 --output-package ${PKG_BASE}/client/clientset --go-header-file hack/boilerplate.go.txt -o ${FAKE_GOPATH}/src

"${BINDIR}/lister-gen" -v 5 -i ${PKG_BASE}/api/machinelearning/v1alpha2 --output-package ${PKG_BASE}/client/listers --go-header-file hack/boilerplate.go.txt -o ${FAKE_GOPATH}/src

"${BINDIR}/informer-gen" -v 5 \
     -i ${PKG_BASE}/api/machinelearning/v1alpha2 \
     --versioned-clientset-package "${PKG_BASE}/client/clientset/versioned" \
     --listers-package "${PKG_BASE}/client/listers" \
     --output-package ${PKG_BASE}/client/informers \
     --go-header-file hack/boilerplate.go.txt \
     -o ${FAKE_GOPATH}/src

export GO111MODULE="on"
cd $REPO_ROOT

# gofmt the tree
#find . -name "*.go" -type f -print0 | xargs -0 gofmt -s -w

