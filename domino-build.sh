#!/bin/bash
set -o nounset -o errexit -o pipefail

nopush_flag=''
while getopts 'n' flag; do
  case "${flag}" in
    n) nopush_flag='true' ;;
    *) error "Unexpected option ${flag}" ;;
  esac
done
readonly nopush_flag

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

if [ "${nopush_flag}" == "" ]; then

  if [ "$(cat ~/.docker/config.json | jq '.auths | has("quay.io")')" = "true" ]; then
    echo -e "[Docker is already logged into quay.io, using existing credentials.]"
  elif [ "${QUAY_USER:-missing}" != "missing" ] && [ "${QUAY_PASSWORD:-missing}" != "missing" ]; then
    echo "$QUAY_PASSWORD" | docker login -u "$QUAY_USER" --password-stdin quay.io
  else
    echo -e "[Push to quay.io requires docker login, either run 'docker login quay.io' or set QUAY_USER and QUAY_PASSWORD before running this script.]"
    exit 1
  fi

  echo -e "\n  Pushing operator...\n"
  docker push "quay.io/domino/seldon-core-operator:${TARGET_IMAGE_TAG}"

  echo -e "\n  Pushing executor...\n"
  docker push "quay.io/domino/seldon-core-executor:${TARGET_IMAGE_TAG}"

fi