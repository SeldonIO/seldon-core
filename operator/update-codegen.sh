#!/usr/bin/env bash

# Generates custom k8s client for our CRs including watchers

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR="$(pwd)"
MODULE="github.com/seldonio/seldon-core/operator/v2"

CODEGEN_PKG="${CODEGEN_PKG:-$(cd "${ROOT_DIR}" && go list -f '{{.Dir}}' -m k8s.io/code-generator)}"

source "${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_client \
    --with-watch \
    --output-dir "${ROOT_DIR}/pkg/generated" \
    --output-pkg "${MODULE}/pkg/generated" \
    --boilerplate "${ROOT_DIR}/hack/boilerplate.go.txt" \
    "${ROOT_DIR}/apis"
