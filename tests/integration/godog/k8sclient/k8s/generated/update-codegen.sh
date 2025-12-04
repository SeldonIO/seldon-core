#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
MODULE="github.com/yourorg/myproject"  # Change to your module name

# Find code-generator location
CODEGEN_PKG="${CODEGEN_PKG:-$(cd "${ROOT_DIR}" && go list -f '{{.Dir}}' -m k8s.io/code-generator)}"

source "${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_helpers \
    --boilerplate "${ROOT_DIR}/hack/boilerplate.go.txt" \
    "${ROOT_DIR}/pkg/apis"

kube::codegen::gen_client \
    --with-watch \
    --output-dir "${ROOT_DIR}/pkg/generated" \
    --output-pkg "${MODULE}/pkg/generated" \
    --boilerplate "${ROOT_DIR}/hack/boilerplate.go.txt" \
    "${ROOT_DIR}/pkg/apis"