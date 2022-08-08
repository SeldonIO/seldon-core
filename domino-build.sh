#!/bin/bash
set -o nounset -o errexit -o pipefail

SELDON_REPO=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "${SELDON_REPO}"

SOURCE_IMAGE_TAG="$(<version.txt)"
TARGET_IMAGE_TAG="${SOURCE_IMAGE_TAG}-$(git rev-parse --short HEAD)"

echo -e "\n  Building operator...\n"
cd "$SELDON_REPO/operator"
make docker-build-no-test

echo -e "\n  Building executor...\n"
cd "$SELDON_REPO/executor"
make docker-build

cd "${SELDON_REPO}"

echo -e "\n  Tagging operator...\n"
docker tag "seldonio/seldon-core-operator:${SOURCE_IMAGE_TAG}" "quay.io/domino/seldon-core-operator:${TARGET_IMAGE_TAG}"

echo -e "\n  Tagging executor...\n"
docker tag "seldonio/seldon-core-executor:${SOURCE_IMAGE_TAG}" "quay.io/domino/seldon-core-executor:${TARGET_IMAGE_TAG}"

echo -e "\n  Pushing operator...\n"
docker push "quay.io/domino/seldon-core-operator:${TARGET_IMAGE_TAG}"

echo -e "\n  Pushing executor...\n"
docker push "quay.io/domino/seldon-core-executor:${TARGET_IMAGE_TAG}"
